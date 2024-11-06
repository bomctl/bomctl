// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch_test.go
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
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type gitFetchSuite struct {
	suite.Suite
	tmpDir string
	*options.Options
	*db.Backend
	*git.Client
	*httptest.Server
	documentInfo []testutil.DocumentInfo
}

func (gfs *gitFetchSuite) setupGitServer() {
	gfs.T().Helper()

	// Create server root and test repository directories.
	serverRoot := filepath.Join(gfs.tmpDir, "git-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo.git")
	gfs.Require().NoError(os.MkdirAll(repoDir, os.ModePerm))

	// Create storage for test Git server repository.
	repoFS := osfs.New(repoDir)
	storer := filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault())

	// Initialize test Git server repository.
	repo, err := gogit.InitWithOptions(storer, repoFS, gogit.InitOptions{DefaultBranch: plumbing.Main})
	gfs.Require().NoError(err)

	worktree, err := repo.Worktree()
	gfs.Require().NoError(err)

	sbomDir := filepath.Join("path", "to")

	gfs.Require().NoError(repoFS.MkdirAll(sbomDir, os.ModePerm))

	testdataDir := testutil.GetTestdataDir()

	entries, err := os.ReadDir(testdataDir)
	gfs.Require().NoError(err)

	for _, entry := range entries {
		targetPath := filepath.Join(sbomDir, entry.Name())
		file, err := repoFS.Create(targetPath)
		gfs.Require().NoError(err)

		content, err := os.ReadFile(filepath.Join(testdataDir, entry.Name()))
		gfs.Require().NoError(err)

		_, err = file.Write(content)
		gfs.Require().NoError(err)

		_, err = worktree.Add(targetPath)
		gfs.Require().NoError(err)
	}

	// Create initial commit and pack Git objects.
	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "bomctl-unit-test",
			Email: "bomctl-unit-test@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	gfs.Require().NoError(err)
	gfs.Require().NoError(repo.Storer.PackRefs())
	gfs.Require().NoError(serverinfo.UpdateServerInfo(storer, repoFS))

	// Update test Git server repository config to bare and unset worktree.
	// This is to allow clients to update the server repository's main branch.
	repoConfig, err := repo.Config()
	gfs.Require().NoError(err)

	repoConfig.Core.IsBare = true
	repoConfig.Core.Worktree = ""

	gfs.Require().NoError(repo.SetConfig(repoConfig), "Failed to set Git repo config")

	// Get path to git executable.
	gitPath, err := exec.LookPath("git")
	gfs.Require().NoError(err, "Unable to find git executable")

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
	gfs.Server = httptest.NewServer(gitHandler)
}

func (gfs *gitFetchSuite) SetupSuite() {
	var err error

	gfs.tmpDir, err = os.MkdirTemp("", "git-fetch-test")
	gfs.Require().NoError(err, "Failed to create temporary directory")

	gfs.Backend, err = testutil.NewTestBackend()
	gfs.Require().NoError(err, "failed database backend creation")

	gfs.documentInfo, err = testutil.AddTestDocuments(gfs.Backend)
	gfs.Require().NoError(err, "failed database backend setup")

	gfs.Client = &git.Client{}

	gfs.setupGitServer()

	gfs.Options = options.New().WithContext(context.WithValue(context.Background(), db.BackendKey{}, gfs.Backend))
}

func (gfs *gitFetchSuite) TearDownSuite() {
	gfs.Server.Close()
	gfs.Backend.CloseClient()

	if err := os.RemoveAll(gfs.tmpDir); err != nil {
		gfs.T().Fatalf("Error removing temp directory %s", gfs.tmpDir)
	}
}

func (gfs *gitFetchSuite) TestClient_Fetch() {
	for _, alias := range []string{"cdx", "spdx"} {
		want, err := os.ReadFile(filepath.Join(testutil.GetTestdataDir(), fmt.Sprintf("sbom.%s.json", alias)))
		gfs.Require().NoError(err)

		gfs.Run(alias, func() {
			opts := &options.FetchOptions{Options: gfs.Options}

			fetchURL := fmt.Sprintf("%s/test/repo.git@main#path/to/sbom.%s.json", gfs.Server.URL, alias)

			got, err := gfs.Fetch(fetchURL, opts)
			gfs.Require().NoError(err)

			gfs.Len(got, len(want))

			document, err := gfs.GetDocumentByIDOrAlias(alias)
			gfs.Require().NoError(err)

			srcData, err := gfs.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.SourceDataAnnotation)
			gfs.Require().NoError(err)

			gfs.Len(srcData, len(want))
		})
	}
}

func TestGitFetchSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitFetchSuite))
}
