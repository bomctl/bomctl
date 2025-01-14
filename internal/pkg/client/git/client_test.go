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
	"net/http/cgi"
	"net/http/httptest"
	neturl "net/url"
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
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type gitClientSuite struct {
	suite.Suite
	tmpDir string
	*options.Options
	*db.Backend
	*git.Client
	*httptest.Server
	documents    []*sbom.Document
	documentInfo []testutil.DocumentInfo
}

func (gcs *gitClientSuite) setupGitServer() {
	gcs.T().Helper()

	// Create server root and test repository directories.
	serverRoot := filepath.Join(gcs.tmpDir, "git-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo.git")
	gcs.Require().NoError(os.MkdirAll(repoDir, os.ModePerm))

	// Create storage for test Git server repository.
	repoFS := osfs.New(repoDir)
	storer := filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault())

	// Initialize test Git server repository.
	repo, err := gogit.InitWithOptions(storer, repoFS, gogit.InitOptions{DefaultBranch: plumbing.Main})
	gcs.Require().NoError(err)

	worktree, err := repo.Worktree()
	gcs.Require().NoError(err)

	sbomDir := filepath.Join("path", "to")

	gcs.Require().NoError(repoFS.MkdirAll(sbomDir, os.ModePerm))

	testdataDir := testutil.GetTestdataDir()

	entries, err := os.ReadDir(testdataDir)
	gcs.Require().NoError(err)

	for _, entry := range entries {
		targetPath := filepath.Join(sbomDir, entry.Name())
		file, err := repoFS.Create(targetPath)
		gcs.Require().NoError(err)

		content, err := os.ReadFile(filepath.Join(testdataDir, entry.Name()))
		gcs.Require().NoError(err)

		_, err = file.Write(content)
		gcs.Require().NoError(err)

		_, err = worktree.Add(targetPath)
		gcs.Require().NoError(err)
	}

	// Create initial commit and pack Git objects.
	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "bomctl-unit-test",
			Email: "bomctl-unit-test@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	gcs.Require().NoError(err)
	gcs.Require().NoError(repo.Storer.PackRefs())
	gcs.Require().NoError(serverinfo.UpdateServerInfo(storer, repoFS))

	// Update test Git server repository config to bare and unset worktree.
	// This is to allow clients to update the server repository's main branch.
	repoConfig, err := repo.Config()
	gcs.Require().NoError(err)

	repoConfig.Core.IsBare = true
	repoConfig.Core.Worktree = ""

	gcs.Require().NoError(repo.SetConfig(repoConfig), "Failed to set Git repo config")

	// Get path to git executable.
	gitPath, err := exec.LookPath("git")
	gcs.Require().NoError(err, "Unable to find git executable")

	// Create CGI handler to handle Git smart protocol requests.
	gitHandler := &cgi.Handler{
		Path: gitPath,
		Args: []string{
			"-c", "http.getanyfile",
			"-c", "http.receivepack",
			"-c", "http.uploadpack",
			"http-backend",
		},
		Env: []string{"GIT_PROJECT_ROOT=" + serverRoot, "GIT_HTTP_EXPORT_ALL=true"},
	}

	// Start the test server.
	gcs.Server = httptest.NewServer(gitHandler)
}

func (gcs *gitClientSuite) SetupSuite() {
	var err error

	gcs.tmpDir, err = os.MkdirTemp("", "git-push-test")
	gcs.Require().NoError(err, "Failed to create temporary directory")

	gcs.Backend, err = testutil.NewTestBackend()
	gcs.Require().NoError(err, "failed database backend creation")

	gcs.documentInfo, err = testutil.AddTestDocuments(gcs.Backend)
	gcs.Require().NoError(err, "failed database backend setup")

	gcs.Client = &git.Client{}

	gcs.setupGitServer()

	for idx := range gcs.documentInfo {
		gcs.documents = append(gcs.documents, gcs.documentInfo[idx].Document)
	}

	gcs.Options = options.New().
		WithCacheDir(gcs.tmpDir).
		WithContext(context.WithValue(context.Background(), db.BackendKey{}, gcs.Backend))
}

func (gcs *gitClientSuite) TearDownSuite() {
	gcs.Server.Close()
	gcs.Backend.CloseClient()

	if err := os.RemoveAll(gcs.tmpDir); err != nil {
		gcs.T().Fatalf("Error removing temp directory %s", gcs.tmpDir)
	}
}

func (gcs *gitClientSuite) TestClient_Init() {
	for _, data := range []struct {
		expected *git.Client
		name     string
		url      string
	}{
		{
			name:     "http scheme",
			url:      "http://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "https scheme with username, port",
			url:      "https://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "https scheme with username, password, port",
			url:      "https://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "https scheme with username",
			url:      "https://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "ssh scheme",
			url:      "ssh://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "ssh scheme with username, port",
			url:      "ssh://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "ssh scheme with username, password, port",
			url:      "ssh://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "ssh scheme with username",
			url:      "ssh://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "git scheme",
			url:      "git://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "git scheme with username, port",
			url:      "git://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "git scheme with username, password, port",
			url:      "git://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "git scheme with username",
			url:      "git://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: &git.Client{},
		},
		{
			name:     "path does not end in .git",
			url:      "https://github.com/bomctl/bomctl?ref=main#sbom.cdx.json",
			expected: nil,
		},
		{
			name:     "missing git ref",
			url:      "https://github.com/bomctl/bomctl.git#sbom.cdx.json",
			expected: nil,
		},
		{
			name:     "missing path to SBOM file",
			url:      "https://github.com/bomctl/bomctl.git?ref=main",
			expected: nil,
		},
	} {
		gcs.Run(data.name, func() {
			testURL, err := neturl.Parse(data.url)
			gcs.Require().NoError(err)

			actual, err := git.Init(testURL)
			if data.expected != nil {
				gcs.Require().NoError(err)
			}

			gcs.Require().Equal(data.expected, actual, data.url)
		})
	}
}

func TestGitClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitClientSuite))
}
