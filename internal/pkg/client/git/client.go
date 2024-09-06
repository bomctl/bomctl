// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// ------------------------------------------------------------------------
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
// ------------------------------------------------------------------------
package git

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

type Client struct {
	repo     *git.Repository
	worktree *git.Worktree
	basePath string
	tmpDir   string
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

func (client *Client) cloneRepo(parsedURL *url.ParsedURL, auth *url.BasicAuth, opts *options.Options) (err error) {
	client.basePath = parsedURL.Fragment

	// Copy parsedRepoURL, excluding auth, git ref, and fragment.
	baseURL := &url.ParsedURL{
		Scheme:   parsedURL.Scheme,
		Hostname: parsedURL.Hostname,
		Path:     parsedURL.Path,
		Port:     parsedURL.Port,
	}

	refName := plumbing.NewBranchReferenceName(parsedURL.GitRef)

	cloneOpts := &git.CloneOptions{
		URL:           baseURL.String(),
		Auth:          auth,
		RemoteName:    "origin",
		ReferenceName: refName,
		SingleBranch:  true,
		Depth:         1,
	}

	opts.Logger.Debug("Cloning git repo", "url", baseURL)

	// Create temp directory to clone into.
	if client.tmpDir, err = os.MkdirTemp(os.TempDir(), strings.ReplaceAll(parsedURL.Path, "/", "-")); err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	// Create any remaining path components.
	if err = os.MkdirAll(filepath.Join(client.tmpDir, filepath.Dir(client.basePath)), fs.ModePerm); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Clone the repository into the temp directory
	if client.repo, err = git.PlainClone(client.tmpDir, false, cloneOpts); err != nil {
		return fmt.Errorf("cloning Git repository: %w", err)
	}

	if client.worktree, err = client.repo.Worktree(); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	return nil
}

func (client *Client) removeTmpDir() {
	defer os.RemoveAll(client.tmpDir)
	client.tmpDir = ""
}
