// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch_test.go
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
package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

var TestDataDir = filepath.Join("..", "..", "db", "testdata")

type gitSuite struct {
	suite.Suite
	tempDir string
	repo    *gogit.Repository
	opts    *options.Options
	backend *db.Backend
	gc      *git.Client
	doc     *sbom.Document
	docs    []*sbom.Document
}

func (gs *gitSuite) SetupSuite() {
	dir, err := os.MkdirTemp("", "testrepo")
	if err != nil {
		gs.T().Fatalf("failed to create temporary directory: %v", err)
	}

	gs.tempDir = dir

	r, err := gogit.PlainInit(dir, false)
	if err != nil {
		gs.T().Fatalf("failed to initialize git repo: %v", err)
	}

	gs.repo = r

	sboms, err := os.ReadDir(TestDataDir)
	if err != nil {
		gs.T().Fatalf("%v", err)
	}

	rdr := reader.New()
	for sbomIdx := range sboms {
		doc, err := rdr.ParseFile(filepath.Join(TestDataDir, sboms[sbomIdx].Name()))
		if err != nil {
			gs.T().Fatalf("%v", err)
		}

		gs.docs = append(gs.docs, doc)
	}

	gs.opts = options.New().
		WithCacheDir(viper.GetString("cache_dir"))

	gs.backend, err = db.NewBackend(
		db.WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)),
	)
	if err != nil {
		gs.T().Fatalf("%v", err)
	}

	for _, document := range gs.docs {
		err := gs.backend.AddDocument(document)
		if err != nil {
			gs.Fail("failed retrieving document", "id", document.GetMetadata().GetId())
		}
	}

	gs.opts = gs.opts.WithContext(context.WithValue(context.Background(), db.BackendKey{}, gs.backend))

	gs.gc = &git.Client{}
}

func (gs *gitSuite) TearDownSuite() {
	err := os.RemoveAll(gs.tempDir)
	if err != nil {
		gs.T().Logf("Error removing repo file %s", db.DatabaseFile)
	}

	gs.backend.CloseClient()

	if _, err := os.Stat(db.DatabaseFile); err == nil {
		if err := os.Remove(db.DatabaseFile); err != nil {
			gs.T().Logf("Error removing database file %s", db.DatabaseFile)
		}
	}
}

func TestGitSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitSuite))
}

func (gs *gitSuite) TestParse() {
	gs.T().Parallel()

	for _, data := range []struct {
		expected *url.ParsedURL
		name     string
		url      string
	}{
		{
			name: "git+http scheme",
			url:  "git+http://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "http",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git+https scheme with username, port",
			url:  "git+https://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "https",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git+https scheme with username, password, port",
			url:  "git+https://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "https",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git+https scheme with username",
			url:  "git+https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "https",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "ssh scheme",
			url:  "ssh://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "ssh scheme with username, port",
			url:  "ssh://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "ssh scheme with username, password, port",
			url:  "ssh://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "ssh scheme with username",
			url:  "ssh://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git scheme",
			url:  "git://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git scheme with username, port",
			url:  "git://git@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git scheme with username, password, port",
			url:  "git://username:password@github.com:12345/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git scheme with username",
			url:  "git://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git SCP-like syntax",
			url:  "github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git SCP-like syntax with username",
			url:  "git@github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git SCP-like syntax with username, password",
			url:  "username:password@github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git SCP-like syntax with username",
			url:  "git@github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name:     "path does not end in .git",
			url:      "git+https://github.com/bomctl/bomctl@main#sbom.cdx.json",
			expected: nil,
		},
		{
			name:     "missing git ref",
			url:      "git+https://github.com/bomctl/bomctl.git#sbom.cdx.json",
			expected: nil,
		},
		{
			name:     "missing path to SBOM file",
			url:      "git+https://github.com/bomctl/bomctl.git@main",
			expected: nil,
		},
	} {
		actual := gs.gc.Parse(data.url)

		gs.Equal(data.expected, actual)
	}
}
