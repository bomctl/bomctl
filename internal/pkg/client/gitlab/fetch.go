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

	"github.com/xanzy/go-gitlab"

	bomctloptions "github.com/bomctl/bomctl/internal/pkg/options"
)

type ClientWrapperInterface interface {
	GetProject(
		any,
		*gitlab.GetProjectOptions,
		...gitlab.RequestOptionFunc,
	) (*gitlab.Project, *gitlab.Response, error)
	GetBranch(any, string, ...gitlab.RequestOptionFunc) (*gitlab.Branch, *gitlab.Response, error)
	GetCommit(
		any,
		string,
		*gitlab.GetCommitOptions,
		...gitlab.RequestOptionFunc,
	) (*gitlab.Commit, *gitlab.Response, error)
	CreateDependencyListExport(
		int,
		*gitlab.CreateDependencyListExportOptions,
		...gitlab.RequestOptionFunc,
	) (*gitlab.DependencyListExport, *gitlab.Response, error)
	GetDependencyListExport(int, ...gitlab.RequestOptionFunc) (*gitlab.DependencyListExport, *gitlab.Response, error)
	DownloadDependencyListExport(int, ...gitlab.RequestOptionFunc) (io.Reader, *gitlab.Response, error)
}

type clientWrapper struct {
	gitlab.ProjectsService
	gitlab.BranchesService
	gitlab.CommitsService
	gitlab.DependencyListExportService
}

type dependencyListExportSession struct {
	Client ClientWrapperInterface
	Export *gitlab.DependencyListExport
}

var (
	errInvalidGitLabURL = errors.New("invalid URL for GitLab fetching")
	errFailedWebRequest = errors.New("web request failed")
	errForbiddenAccess  = errors.New("you don't have permission to read the dependency list")
)

func newDependencyListExportSession(baseURL, gitLabToken string) (*dependencyListExportSession, error) {
	gitlabClient, err := gitlab.NewClient(gitLabToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	newSession := &dependencyListExportSession{
		Client: &clientWrapper{
			ProjectsService:             *gitlabClient.Projects,
			BranchesService:             *gitlabClient.Branches,
			CommitsService:              *gitlabClient.Commits,
			DependencyListExportService: *gitlabClient.DependencyListExport,
		},
		Export: nil,
	}

	return newSession, nil
}

func validateHTTPStatusCode(statusCode int) error {
	if statusCode < http.StatusOK || http.StatusMultipleChoices <= statusCode {
		return fmt.Errorf("%w. HTTP status code: %d", errFailedWebRequest, statusCode)
	}

	return nil
}

func (session *dependencyListExportSession) createExport(projectName, branchName string) error {
	project, response, err := session.Client.GetProject(projectName, nil)
	if err != nil {
		return fmt.Errorf("failed to get project info: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	branch, response, err := session.Client.GetBranch(project.ID, branchName)
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	commit, response, err := session.Client.GetCommit(project.ID, branch.Commit.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	// NOTE:
	// If an authenticated user does not have permission to read_dependency,
	// this request returns a 403 Forbidden status code.
	export, response, err := session.Client.CreateDependencyListExport(commit.LastPipeline.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to create dependency list export: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	if response.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w", errForbiddenAccess)
	}

	session.Export = export

	return nil
}

func (session *dependencyListExportSession) pollExportUntilFinished() error {
	const waitSeconds = 2
	for !session.Export.HasFinished {
		time.Sleep(waitSeconds * time.Second)

		updatedExport, response, err := session.Client.GetDependencyListExport(session.Export.ID)
		if err != nil {
			return fmt.Errorf("failed to get dependency list export: %w", err)
		}

		if err := validateHTTPStatusCode(response.StatusCode); err != nil {
			return err
		}

		session.Export = updatedExport
	}

	return nil
}

func (session *dependencyListExportSession) downloadExport() ([]byte, error) {
	sbomReader, response, err := session.Client.DownloadDependencyListExport(session.Export.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to download dependency list: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return nil, err
	}

	var sbomData []byte

	_, err = sbomReader.Read(sbomData)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM data: %w", err)
	}

	return sbomData, nil
}

func (client *Client) Fetch(fetchURL string, _ *bomctloptions.FetchOptions) ([]byte, error) {
	url := client.Parse(fetchURL)
	if url == nil {
		return nil, fmt.Errorf("%w: %s", errInvalidGitLabURL, fetchURL)
	}

	domain := url.Hostname
	if url.Port != "" {
		domain = fmt.Sprintf("%s:%s", domain, url.Port)
	}

	baseURL := fmt.Sprintf("https://%s/api/v4", domain)
	projectName := url.Path
	branchName := url.Fragment

	gitLabToken := os.Getenv("GITLAB_FETCH_TOKEN")

	session, err := newDependencyListExportSession(baseURL, gitLabToken)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := session.createExport(projectName, branchName); err != nil {
		return nil, err
	}

	if err := session.pollExportUntilFinished(); err != nil {
		return nil, err
	}

	sbomData, err := session.downloadExport()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sbomData, nil
}
