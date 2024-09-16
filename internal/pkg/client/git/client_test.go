// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/client_test.go
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

package git_test

import (
	"context"
	"fmt"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/serverinfo"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

type gitSuite struct {
	suite.Suite
	tmpDir  string
	opts    *options.Options
	backend *db.Backend
	gc      *git.Client
	server  *httptest.Server
	docs    []*sbom.Document
}

func (gs *gitSuite) setupGitServer() {
	gs.T().Helper()

	// Create server root and test repository directories.
	serverRoot := filepath.Join(gs.tmpDir, "git-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo.git")
	gs.Require().NoError(os.MkdirAll(repoDir, os.ModePerm))

	// Create storage for test Git server repository.
	repoFS := osfs.New(repoDir)
	storer := filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault())

	// Initialize test Git server repository.
	repo, err := gogit.InitWithOptions(storer, repoFS, gogit.InitOptions{DefaultBranch: plumbing.Main})
	gs.Require().NoError(err)

	worktree, err := repo.Worktree()
	gs.Require().NoError(err)

	// Create initial commit and pack Git objects.
	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "bomctl-unit-test",
			Email: "bomctl-unit-test@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	gs.Require().NoError(err)
	gs.Require().NoError(repo.Storer.PackRefs())
	gs.Require().NoError(serverinfo.UpdateServerInfo(storer, repoFS))

	// Update test Git server repository config to bare and unset worktree.
	// This is to allow clients to update the server repository's main branch.
	repoConfig, err := repo.Config()
	gs.Require().NoError(err)

	repoConfig.Core.IsBare = true
	repoConfig.Core.Worktree = ""

	gs.Require().NoError(repo.SetConfig(repoConfig), "Failed to set Git repo config")

	// Get path to git executable.
	gitPath, err := exec.LookPath("git")
	gs.Require().NoError(err, "Unable to find git executable")

	// Create CGI handler to handle Git smart protocol requests.
	gitHandler := &cgi.Handler{
		Path: gitPath,
		Args: []string{
			"-c", "http.getanyfile",
			"-c", "http.receivepack",
			"-c", "http.uploadpack",
			"http-backend",
		},
		Env: []string{fmt.Sprintf("GIT_PROJECT_ROOT=%s", serverRoot), "GIT_HTTP_EXPORT_ALL=true"},
	}

	// Start the test server.
	gs.server = httptest.NewServer(gitHandler)
}

func (gs *gitSuite) BeforeTest(_suiteName, _testName string) {
	var err error

	gs.tmpDir, err = os.MkdirTemp("", "git-client-test")
	gs.Require().NoErrorf(err, "Failed to create temporary directory: %v", err)

	gs.setupGitServer()

	pushOpts := &options.PushOptions{Options: gs.opts}
	gs.Require().NoError(
		gs.gc.PreparePush(
			fmt.Sprintf("%s/test/repo.git@main#path/to/sbom.cdx", gs.server.URL),
			pushOpts,
		),
	)

	repoConfig, err := gs.gc.Repo().Config()
	gs.Require().NoError(err)

	repoConfig.Author.Name = "bomctl-unit-test"
	repoConfig.Author.Email = "bomctl-unit-test@users.noreply.github.com"

	gs.Require().NoError(gs.gc.Repo().SetConfig(repoConfig), "Failed to set Git repo config")
}

func (gs *gitSuite) AfterTest(_suiteName, _testName string) {
	gs.server.Close()

	defer os.RemoveAll(gs.tmpDir)
}

func (gs *gitSuite) SetupSuite() {
	backend, err := db.NewBackend(db.WithDatabaseFile(db.DatabaseFile))
	if err != nil {
		gs.T().Fatalf("%v", err)
	}

	gs.backend = backend

	testdataDir := filepath.Join("..", "..", "db", "testdata")

	sboms, err := os.ReadDir(testdataDir)
	if err != nil {
		gs.T().Fatalf("%v", err)
	}

	for sbomIdx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join(testdataDir, sboms[sbomIdx].Name()))
		if err != nil {
			gs.T().Fatalf("%v", err)
		}

		doc, err := gs.backend.AddDocument(sbomData)
		if err != nil {
			gs.FailNow("failed storing document", "err", err)
		}

		gs.docs = append(gs.docs, doc)
	}

	gs.opts = options.New().WithCacheDir(viper.GetString("cache_dir"))
	gs.opts = gs.opts.WithContext(context.WithValue(context.Background(), db.BackendKey{}, gs.backend))
	gs.gc = &git.Client{}
}

func (gs *gitSuite) TearDownSuite() {
	gs.server.Close()

	err := os.RemoveAll(gs.tmpDir)
	if err != nil {
		gs.T().Fatalf("Error removing repo file %s", db.DatabaseFile)
	}

	gs.backend.CloseClient()

	if _, err := os.Stat(db.DatabaseFile); err == nil {
		if err := os.Remove(db.DatabaseFile); err != nil {
			gs.T().Fatalf("Error removing database file %s", db.DatabaseFile)
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
