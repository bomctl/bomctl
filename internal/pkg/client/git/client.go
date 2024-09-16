// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/client.go
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

package git

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

type Client struct{}

func (*Client) Name() string {
	return "Git"
}

func (*Client) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^%s%s%s%s%s$",
			`((?:git\+)?(?P<scheme>https?|git|ssh):\/\/)?`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`(?P<hostname>[^@\/?#:]+)(?::(?P<port>\d+))?`,
			`(?:[\/:](?P<path>[^@#]+\.git)@?)`,
			`((?:@(?P<gitRef>[^#]+))(?:#(?P<fragment>.*)))?`,
		),
	)
}

func (client *Client) Parse(rawURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := client.RegExp()
	match := pattern.FindStringSubmatch(rawURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	if results["scheme"] == "" {
		results["scheme"] = "ssh"
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname", "path", "gitRef", "fragment"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	return &url.ParsedURL{
		Scheme:   results["scheme"],
		Username: results["username"],
		Password: results["password"],
		Hostname: results["hostname"],
		Port:     results["port"],
		Path:     results["path"],
		GitRef:   results["gitRef"],
		Query:    results["query"],
		Fragment: results["fragment"],
	}
}

func cloneRepo(tempDir string, parsedRepoURL *url.ParsedURL, auth *url.BasicAuth,
	opts *options.Options,
) (*git.Repository, error) {
	refName := plumbing.NewBranchReferenceName(parsedRepoURL.GitRef)

	// Copy parsedRepoURL, excluding auth, git ref, and fragment.
	baseURL := &url.ParsedURL{
		Scheme:   parsedRepoURL.Scheme,
		Hostname: parsedRepoURL.Hostname,
		Path:     parsedRepoURL.Path,
		Port:     parsedRepoURL.Port,
	}

	cloneOpts := &git.CloneOptions{
		URL:           baseURL.String(),
		Auth:          auth,
		RemoteName:    "origin",
		ReferenceName: refName,
		SingleBranch:  true,
		Depth:         1,
	}

	opts.Logger.Debug("Cloning git repo: %s", baseURL)
	// Clone the repository into the temp directory
	repo, err := git.PlainClone(tempDir, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to clone Git repository: %w", err)
	}

	return repo, nil
}
