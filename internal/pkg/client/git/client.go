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

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type Client struct {
	repo     *git.Repository
	worktree *git.Worktree
	basePath string
}

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

func (client *Client) Parse(rawURL string) *netutil.URL {
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

	return &netutil.URL{
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

func (client *Client) cloneRepo(url *netutil.URL, auth *netutil.BasicAuth, opts *options.Options) (err error) {
	client.basePath = url.Fragment

	// Copy parsedRepoURL, excluding auth, git ref, and fragment.
	baseURL := &netutil.URL{
		Scheme:   url.Scheme,
		Hostname: url.Hostname,
		Path:     url.Path,
		Port:     url.Port,
	}

	cloneOpts := &git.CloneOptions{
		URL:           baseURL.String(),
		Auth:          auth,
		RemoteName:    git.DefaultRemoteName,
		ReferenceName: plumbing.NewBranchReferenceName(url.GitRef),
		SingleBranch:  true,
		Depth:         1,
	}

	opts.Logger.Debug("Cloning git repo", "url", baseURL)

	if client.repo, err = git.Clone(memory.NewStorage(), memfs.New(), cloneOpts); err != nil {
		return fmt.Errorf("cloning Git repository: %w", err)
	}

	if client.worktree, err = client.repo.Worktree(); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	return nil
}
