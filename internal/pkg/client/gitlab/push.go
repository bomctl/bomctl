// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/gitlab/push.go
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
	"os"
	"regexp"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/outpututil"
)

type stringWriter struct {
	*strings.Builder
}

var (
	errInvalidSbomID      = errors.New("invalid SBOM ID")
	errMissingPackageInfo = errors.New("missing package name or version")
)

func (*stringWriter) Close() error {
	return nil
}

func (client *Client) PreparePush(pushURL string, _opts *options.PushOptions) error {
	gitLabToken := os.Getenv("BOMCTL_GITLAB_TOKEN")

	url := client.Parse(pushURL)

	host := url.Hostname

	if url.Port != "" {
		host = fmt.Sprintf("%s:%s", host, url.Port)
	}

	baseURL := fmt.Sprintf("https://%s/api/v4", host)

	gitLabClient, err := gitlab.NewClient(gitLabToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	client.GitLabToken = gitLabToken
	client.projectProvider = gitLabClient.Projects
	client.genericPackagePublisher = gitLabClient.GenericPackages

	client.PushQueue = make([]*sbomFile, 0)

	return nil
}

func (client *Client) AddFile(_pushURL, id string, opts *options.PushOptions) error {
	opts.Logger.Info("Adding file", "id", id)

	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	sbom, err := backend.GetDocumentByIDOrAlias(id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	sbomFormat := sbom.GetMetadata().GetSourceData().GetFormat()

	isCycloneDX := strings.Contains(sbomFormat, "cyclonedx")

	var uuidRegex *regexp.Regexp

	if isCycloneDX {
		uuidRegex = regexp.MustCompile(`^urn:uuid:([\w-]+)$`)
	} else {
		uuidRegex = regexp.MustCompile(`^.+/([^/#]+)(?:#\w+)?`)
	}

	uuidMatch := uuidRegex.FindStringSubmatch(id)

	if len(uuidMatch) == 0 {
		return fmt.Errorf("%w: %s", errInvalidSbomID, id)
	}

	sbomFilename := uuidMatch[1]

	xmlFormatRegex := regexp.MustCompile(`\bxml\b`)
	xmlFormatMatch := xmlFormatRegex.FindStringSubmatch(string(opts.Format))

	if len(xmlFormatMatch) == 0 {
		sbomFilename += ".json"
	} else {
		sbomFilename += ".xml"
	}

	sbomWriter := &stringWriter{&strings.Builder{}}
	if err := outpututil.WriteStream(sbom, opts.Format, opts.Options, sbomWriter); err != nil {
		return fmt.Errorf("failed to serialize SBOM %s: %w", id, err)
	}

	client.PushQueue = append(client.PushQueue, &sbomFile{
		Name:     sbomFilename,
		Contents: sbomWriter.String(),
	})

	return nil
}

func (client *Client) Push(pushURL string, _opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	if url == nil {
		return fmt.Errorf("%w: %s", errInvalidGitLabURL, pushURL)
	}

	project, response, err := client.GetProject(url.Path, nil)
	if err != nil {
		return fmt.Errorf("failed to get project info: %w", err)
	}

	if err := validateHTTPStatusCode(response.StatusCode); err != nil {
		return err
	}

	packageInfo := strings.Split(url.Fragment, "@")

	if packageInfoExpectedLength := 2; len(packageInfo) != packageInfoExpectedLength {
		return fmt.Errorf("%w: %s", errMissingPackageInfo, url.Fragment)
	}

	packageName := packageInfo[0]
	packageVersion := packageInfo[1]

	for _, sbomFile := range client.PushQueue {
		sbomReader := strings.NewReader(sbomFile.Contents)

		_, response, err := client.genericPackagePublisher.PublishPackageFile(
			project.ID,
			packageName,
			packageVersion,
			sbomFile.Name,
			sbomReader,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to push sbom: %w", err)
		}

		if err := validateHTTPStatusCode(response.StatusCode); err != nil {
			return err
		}
	}

	return nil
}
