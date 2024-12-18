// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/gitlab/fetch.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// -----------------------------------------------------------------------------

package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type (
	ProjectProvider interface {
		GetProject(
			any,
			*gitlab.GetProjectOptions,
			...gitlab.RequestOptionFunc,
		) (*gitlab.Project, *gitlab.Response, error)
	}

	BranchProvider interface {
		GetBranch(
			any,
			string,
			...gitlab.RequestOptionFunc,
		) (*gitlab.Branch, *gitlab.Response, error)
	}

	CommitProvider interface {
		GetCommit(
			any,
			string,
			*gitlab.GetCommitOptions,
			...gitlab.RequestOptionFunc,
		) (*gitlab.Commit, *gitlab.Response, error)
	}

	DependencyListExporter interface {
		CreateDependencyListExport(
			int,
			*gitlab.CreateDependencyListExportOptions,
			...gitlab.RequestOptionFunc,
		) (*gitlab.DependencyListExport, *gitlab.Response, error)
		GetDependencyListExport(
			int,
			...gitlab.RequestOptionFunc,
		) (*gitlab.DependencyListExport, *gitlab.Response, error)
		DownloadDependencyListExport(int, ...gitlab.RequestOptionFunc) (io.Reader, *gitlab.Response, error)
	}
)

var (
	errInvalidGitLabURL = errors.New("invalid URL for GitLab fetching")
	errFailedWebRequest = errors.New("web request failed")
	errForbiddenAccess  = errors.New("the supplied token is missing the read_dependency permission")
)

func validateHTTPStatusCode(statusCode int) error {
	if statusCode < http.StatusOK || http.StatusMultipleChoices <= statusCode {
		return fmt.Errorf("%w. HTTP status code: %d", errFailedWebRequest, statusCode)
	}

	return nil
}

func (client *Client) createExport(projectName, branchName string) error {
	project, response, err := client.GetProject(projectName, nil)
	if err != nil {
		return fmt.Errorf("failed to get project info: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	branch, response, err := client.GetBranch(project.ID, branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	commit, response, err := client.GetCommit(project.ID, branch.Commit.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	// NOTE:
	// If an authenticated user does not have permission to read_dependency,
	// this request returns a 403 Forbidden status code.
	export, response, err := client.CreateDependencyListExport(commit.LastPipeline.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create dependency list export: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	if response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w", errForbiddenAccess)
	}

	client.Export = export

	return nil
}

func (client *Client) pollExportUntilFinished() error {
	const waitSeconds = 2
	for !client.Export.HasFinished {
		time.Sleep(waitSeconds * time.Second)

		updatedExport, response, err := client.GetDependencyListExport(client.Export.ID)
		if err != nil {
			return fmt.Errorf("failed to get dependency list export: %w", err)
		}

		if err := validateHTTPStatusCode(response.StatusCode); err != nil {
			return err
		}

		client.Export = updatedExport
	}

	return nil
}

func (client *Client) downloadExport() ([]byte, error) {
	sbomReader, response, err := client.DownloadDependencyListExport(client.Export.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to download dependency list: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return nil, err
	}

	sbomData, err := io.ReadAll(sbomReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM data: %w", err)
	}

	return sbomData, nil
}

func (client *Client) PrepareFetch(url *netutil.URL, _auth *netutil.BasicAuth, _opts *options.Options) error {
	gitLabToken := os.Getenv("BOMCTL_GITLAB_TOKEN")

	host := url.Hostname

	if url.Port != "" {
		host = fmt.Sprintf("%s:%s", host, url.Port)
	}

	baseURL := fmt.Sprintf("https://%s/api/v4", host)

	gitLabClient, err := gitlab.NewClient(gitLabToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return fmt.Errorf("failed to initialize the client: %w", err)
	}

	client.GitLabToken = gitLabToken
	client.ProjectProvider = gitLabClient.Projects
	client.BranchProvider = gitLabClient.Branches
	client.CommitProvider = gitLabClient.Commits
	client.DependencyListExporter = gitLabClient.DependencyListExport

	return nil
}

func (client *Client) Fetch(fetchURL string, _ *options.FetchOptions) ([]byte, error) {
	url := client.Parse(fetchURL)
	if url == nil {
		return nil, fmt.Errorf("%w: %s", errInvalidGitLabURL, fetchURL)
	}

	if err := client.createExport(url.Path, url.GitRef); err != nil {
		return nil, err
	}

	if err := client.pollExportUntilFinished(); err != nil {
		return nil, err
	}

	sbomData, err := client.downloadExport()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sbomData, nil
}
