// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/push_test.go
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
	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type gitPushSuite struct {
	suite.Suite
	tmpDir string
	*options.Options
	*db.Backend
	*git.Client
	*httptest.Server
	documents    []*sbom.Document
	documentInfo []testutil.DocumentInfo
}

func (gps *gitPushSuite) setupGitServer() {
	gps.T().Helper()

	// Create server root and test repository directories.
	serverRoot := filepath.Join(gps.tmpDir, "git-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo.git")
	gps.Require().NoError(os.MkdirAll(repoDir, os.ModePerm))

	// Create storage for test Git server repository.
	repoFS := osfs.New(repoDir)
	storer := filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault())

	// Initialize test Git server repository.
	repo, err := gogit.InitWithOptions(storer, repoFS, gogit.InitOptions{DefaultBranch: plumbing.Main})
	gps.Require().NoError(err)

	worktree, err := repo.Worktree()
	gps.Require().NoError(err)

	// Create initial commit and pack Git objects.
	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "bomctl-unit-test",
			Email: "bomctl-unit-test@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	gps.Require().NoError(err)
	gps.Require().NoError(repo.Storer.PackRefs())
	gps.Require().NoError(serverinfo.UpdateServerInfo(storer, repoFS))

	// Update test Git server repository config to bare and unset worktree.
	// This is to allow clients to update the server repository's main branch.
	repoConfig, err := repo.Config()
	gps.Require().NoError(err)

	repoConfig.Core.IsBare = true
	repoConfig.Core.Worktree = ""

	gps.Require().NoError(repo.SetConfig(repoConfig), "Failed to set Git repo config")

	// Get path to git executable.
	gitPath, err := exec.LookPath("git")
	gps.Require().NoError(err, "Unable to find git executable")

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
	gps.Server = httptest.NewServer(gitHandler)
}

func (gps *gitPushSuite) SetupSuite() {
	var err error

	gps.tmpDir, err = os.MkdirTemp("", "git-push-test")
	gps.Require().NoError(err, "Failed to create temporary directory")

	gps.Backend, err = testutil.NewTestBackend()
	gps.Require().NoError(err, "failed database backend creation")

	gps.documentInfo, err = testutil.AddTestDocuments(gps.Backend)
	gps.Require().NoError(err, "failed database backend setup")

	gps.Client = &git.Client{}

	gps.setupGitServer()

	for idx := range gps.documentInfo {
		gps.documents = append(gps.documents, gps.documentInfo[idx].Document)
	}

	gps.Options = options.New().
		WithCacheDir(gps.tmpDir).
		WithContext(context.WithValue(context.Background(), db.BackendKey{}, gps.Backend))
}

func (gps *gitPushSuite) BeforeTest(_suiteName, _testName string) {
	var err error

	pushOpts := &options.PushOptions{Options: gps.Options}
	gps.Require().NoError(
		gps.Client.PreparePush(
			gps.Server.URL+"/test/repo.git@main#path/to/sbom.cdx",
			pushOpts,
		),
	)

	repoConfig, err := gps.Client.Repo().Config()
	gps.Require().NoError(err)

	repoConfig.Author.Name = "bomctl-unit-test"
	repoConfig.Author.Email = "bomctl-unit-test@users.noreply.github.com"

	gps.Require().NoError(gps.Client.Repo().SetConfig(repoConfig), "Failed to set Git repo config")
}

func (gps *gitPushSuite) TearDownSuite() {
	gps.Server.Close()
	gps.Backend.CloseClient()

	if err := os.RemoveAll(gps.tmpDir); err != nil {
		gps.T().Fatalf("Error removing temp directory %s", db.DatabaseFile)
	}
}

func (gps *gitPushSuite) TestClient_AddFile() {
	opts := &options.PushOptions{
		Options: gps.Options,
		Format:  formats.CDX15JSON,
	}

	if err := gps.Client.AddFile(
		gps.Server.URL+"/test/repo.git@main#path/to/sbom.cdx.json",
		gps.documents[0].GetMetadata().GetId(),
		opts,
	); err != nil {
		gps.FailNowf("Error testing AddFile", "%s", err.Error())
	}

	// Get worktree status.
	status, err := gps.Client.Worktree().Status()
	gps.Require().NoError(err)

	// File must be staged with a status of "A" (added).
	gps.Require().Equal("A", string(status.File("path/to/sbom.cdx.json").Staging))
}

func (gps *gitPushSuite) TestClient_Push() {
	pushURL := gps.Server.URL + "/test/repo.git@main#path/to/sbom.cdx.json"

	opts := &options.PushOptions{
		Options: gps.Options,
		Format:  formats.CDX15JSON,
		UseTree: false,
	}

	gps.Require().NoError(gps.Client.AddFile(pushURL, gps.documents[0].GetMetadata().GetId(), opts))
	gps.Require().NoError(gps.Client.Push(pushURL, opts))
}

func (gps *gitPushSuite) TestGetDocument() {
	for _, document := range gps.documents {
		retrieved, err := git.GetDocument(document.GetMetadata().GetId(), gps.Options)
		if err != nil {
			gps.FailNow("failed retrieving document", "id", document.GetMetadata().GetId())
		}

		gps.Require().Equal(document.GetMetadata().GetId(), retrieved.GetMetadata().GetId())
		gps.Require().Len(retrieved.GetNodeList().GetNodes(), len(document.GetNodeList().GetNodes()))
		gps.Require().Equal(document.GetNodeList().GetRootElements(), retrieved.GetNodeList().GetRootElements())
	}
}

func TestGitPushSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(gitPushSuite))
}
