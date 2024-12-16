// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/gitlab/client.go
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
	"fmt"
	neturl "net/url"
	"os"
	"regexp"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
)

type Client struct {
	ProjectProvider
	BranchProvider
	CommitProvider
	DependencyListExporter
	Export      *gitlab.DependencyListExport
	GitLabToken string
}

func (client *Client) init(sourceURL string) error {
	gitLabToken := os.Getenv("BOMCTL_GITLAB_TOKEN")

	url, err := neturl.Parse(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to parse the url: %w", err)
	}

	baseURL := fmt.Sprintf("https://%s/api/v4", url.Host)

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

func (*Client) Name() string {
	return "GitLab"
}

func (*Client) RegExp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("(?i)^%s%s%s$",
		`(?P<scheme>https?|git|ssh):\/\/`,
		`(?P<hostname>[^@\/?#:]+gitlab[^@\/?#:]+)(?::(?P<port>\d+))?/`,
		`(?P<path>[^@#]+)@(?P<branch>\S+)`))
}

func (client *Client) Parse(rawURL string) *netutil.URL {
	results := map[string]string{}
	pattern := client.RegExp()
	match := pattern.FindStringSubmatch(rawURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname", "path", "branch"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	if err := client.init(rawURL); err != nil {
		return nil
	}

	return &netutil.URL{
		Scheme:   results["scheme"],
		Hostname: results["hostname"],
		Port:     results["port"],
		Path:     results["path"],
		Fragment: results["branch"],
	}
}
