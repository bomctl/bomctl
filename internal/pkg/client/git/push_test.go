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
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/formats"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type gitPushSuite struct {
	gitClientSuite
}

func (gps *gitPushSuite) BeforeTest(_suiteName, _testName string) {
	var err error

	pushOpts := &options.PushOptions{Options: gps.Options}
	gps.Require().NoError(
		gps.Client.PreparePush(
			gps.Server.URL+"/test/repo.git@main#path/to/sbom.cdx.json",
			pushOpts,
		),
	)

	repoConfig, err := gps.Client.Repo().Config()
	gps.Require().NoError(err)

	repoConfig.Author.Name = "bomctl-unit-test"
	repoConfig.Author.Email = "bomctl-unit-test@users.noreply.github.com"

	gps.Require().NoError(gps.Client.Repo().SetConfig(repoConfig), "Failed to set Git repo config")
}

func (gps *gitPushSuite) TestClient_AddFile() {
	pushURL := gps.Server.URL + "/test/repo.git@main#path/to/testbom.cdx.json"

	opts := &options.PushOptions{
		Options: gps.Options,
		Format:  formats.CDX15JSON,
	}

	gps.Require().NoError(gps.Client.AddFile(pushURL, gps.documents[0].GetMetadata().GetId(), opts))

	// Get worktree status.
	status, err := gps.Client.Worktree().Status()
	gps.Require().NoError(err)

	// File must be staged with a status of "A" (added).
	gps.Require().Equal(gogit.Added, status.File("path/to/testbom.cdx.json").Staging)
}

func (gps *gitPushSuite) TestClient_Push() {
	pushURL := gps.Server.URL + "/test/repo.git@main#path/to/testbom.cdx.json"

	opts := &options.PushOptions{
		Options: gps.Options,
		Format:  formats.CDX15JSON,
	}

	gps.Require().NoError(gps.Client.AddFile(pushURL, gps.documents[0].GetMetadata().GetId(), opts))
	gps.Require().NoError(gps.Client.Push(pushURL, opts))

	// Get worktree status.
	status, err := gps.Client.Worktree().Status()
	gps.Require().NoError(err)
	gps.True(status.IsClean())
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
