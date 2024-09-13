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
	"fmt"

	"github.com/protobom/protobom/pkg/formats"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (gs *gitSuite) TestPush() {
	pushURL := fmt.Sprintf("%s/test/repo.git@main#path/to/sbom.cdx.json", gs.server.URL)

	opts := &options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
		UseTree: false,
	}

	gs.Require().NoError(gs.gc.AddFile(pushURL, gs.docs[0].GetMetadata().GetId(), opts))
	gs.Require().NoError(gs.gc.Push("", pushURL, opts))
}

func (gs *gitSuite) TestAddFile() {
	opts := &options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
	}

	if err := gs.gc.AddFile(
		fmt.Sprintf("%s/test/repo.git@main#path/to/sbom.cdx.json", gs.server.URL),
		gs.docs[0].GetMetadata().GetId(),
		opts,
	); err != nil {
		gs.FailNowf("Error testing AddFile", "%s", err.Error())
	}

	// Get worktree status.
	status, err := gs.gc.Worktree().Status()
	gs.Require().NoError(err)

	// File must be staged with a status of "A" (added).
	gs.Require().Equal("A", string(status.File("path/to/sbom.cdx.json").Staging))
}

func (gs *gitSuite) TestGetDocument() {
	for _, document := range gs.docs {
		retrieved, err := git.GetDocument(document.GetMetadata().GetId(), gs.opts)
		if err != nil {
			gs.FailNow("failed retrieving document", "id", document.GetMetadata().GetId())
		}

		gs.Require().Equal(document.GetMetadata().GetId(), retrieved.GetMetadata().GetId())
		gs.Require().Len(retrieved.GetNodeList().GetNodes(), len(document.GetNodeList().GetNodes()))
		gs.Require().Equal(document.GetNodeList().GetRootElements(), retrieved.GetNodeList().GetRootElements())
	}
}
