// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/push/push_test.go
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

package e2e_push_test

import (
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
	"github.com/rogpeppe/go-internal/testscript"

	"github.com/bomctl/bomctl/cmd"
	"github.com/bomctl/bomctl/internal/e2e/e2eutil"
)

func setupGitServer(t *testing.T, tmpDir string) *httptest.Server {
	t.Helper()
	// Create server root and test repository directories.
	serverRoot := filepath.Join(tmpDir, "git-test-server")
	repoDir := filepath.Join(serverRoot, "test", "repo.git")

	err := os.MkdirAll(repoDir, os.ModePerm)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Create storage for test Git server repository.
	repoFS := osfs.New(repoDir)
	storer := filesystem.NewStorage(repoFS, cache.NewObjectLRUDefault())

	// Initialize test Git server repository.
	repo, err := gogit.InitWithOptions(storer, repoFS, gogit.InitOptions{DefaultBranch: plumbing.Main})
	if err != nil {
		t.Fatalf("%v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Create initial commit and pack Git objects.
	_, err = worktree.Commit("Initial commit", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "bomctl-e2e-test",
			Email: "bomctl-e2e-test@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = repo.Storer.PackRefs()
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = serverinfo.UpdateServerInfo(storer, repoFS)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Update test Git server repository config to bare and unset worktree.
	// This is to allow clients to update the server repository's main branch.
	repoConfig, err := repo.Config()
	if err != nil {
		t.Fatalf("%v", err)
	}

	repoConfig.Core.IsBare = true
	repoConfig.Core.Worktree = ""

	err = repo.SetConfig(repoConfig)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Get path to git executable.
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("%v", err)
	}

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
	return httptest.NewServer(gitHandler)
}

func TestBomctlPush(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:                 ".",
		RequireExplicitExec: true,
		Cmds:                e2eutil.CustomCommands(),
		Setup: func(env *testscript.Env) error {
			server := setupGitServer(t, env.Getenv("WORK"))

			pushURL := server.URL + "/test/repo.git?ref=main#path/to/sbom.cdx.json"
			env.Setenv("PUSH_URL", pushURL)
			env.Setenv("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))

			return nil
		},
	})
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{"bomctl": cmd.Execute}))
}
