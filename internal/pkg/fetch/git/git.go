// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/fetch/git/git.go
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
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/url"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

type Fetcher struct{}

func (fetcher *Fetcher) Name() string {
	return "Git"
}

func (fetcher *Fetcher) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^%s%s%s%s%s$",
			`((?:git\+)?(?P<scheme>https?|git|ssh):\/\/)?`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`((?P<hostname>[^@\/?#:]+))(?::(?P<port>\d+))?`,
			`(?:[\/:](?P<path>[^@#]+\.git)@?)`,
			`((?:@(?P<gitRef>[^#]+))(?:#(?P<fragment>.*)))?`,
		),
	)
}

func (fetcher *Fetcher) Parse(fetchURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := fetcher.RegExp()
	match := pattern.FindStringSubmatch(fetchURL)

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

func (fetcher *Fetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) (*sbom.Document, error) {
	// Create temp directory to clone into
	tmpDir, err := os.MkdirTemp(os.TempDir(), "repo")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tmpDir)

	refName := plumbing.NewBranchReferenceName(parsedURL.GitRef)

	// Copy parsedURL, excluding auth, git ref, and fragment.
	baseURL := &url.ParsedURL{
		Scheme:   parsedURL.Scheme,
		Hostname: parsedURL.Hostname,
		Path:     parsedURL.Path,
		Port:     parsedURL.Port,
	}

	cloneOpts := &git.CloneOptions{
		URL:           baseURL.String(),
		Auth:          auth,
		RemoteName:    "origin",
		ReferenceName: refName,
		SingleBranch:  true,
		Depth:         1,
	}

	// Clone the repository into the temp directory
	_, err = git.PlainClone(tmpDir, false, cloneOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to clone Git repository: %w", err)
	}

	// Read the file specified in the URL fragment
	sbomBytes, err := os.ReadFile(filepath.Join(tmpDir, parsedURL.Fragment))
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", parsedURL.Fragment, err)
	}

	// Parse the file content
	document, err := utils.ParseSBOMData(sbomBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing SBOM file content: %w", err)
	}

	return document, nil
}
