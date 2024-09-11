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
	"io/fs"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/formats"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (gs *gitSuite) TestPush() {
	rawURL := "git+https://git@github.com:12345/test/repo.git@main#sbom.cdx.json"

	opts := &options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
		UseTree: false,
	}

	for _, document := range gs.docs {
		if err := gs.gc.AddFile(rawURL, document.GetMetadata().GetId(), opts); err != nil {
			gs.Assert().NoErrorf(err, "Error staging file: %v", err)
		}
	}

	if err := gs.gc.Push("", rawURL, opts); err != nil {
		gs.Assert().NoError(err, "Error testing Push: %v", err)
	}
}

func (gs *gitSuite) TestAddFile() {
	opts := &options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
	}

	gs.gc.SetCloneFunc(func(string, bool, *gogit.CloneOptions) (*gogit.Repository, error) {
		return gs.repo, nil
	})

	testFile := filepath.Join(gs.tmpDir, "test", "sbom.cdx.json")

	if err := os.MkdirAll(filepath.Dir(testFile), fs.ModePerm); err != nil {
		gs.FailNow("failed creating directory")
	}

	if err := gs.gc.AddFile(
		"git+https://git@github.com:12345/test/repo.git@main#test/sbom.cdx.json",
		gs.docs[0].GetMetadata().GetId(),
		opts,
	); err != nil {
		gs.FailNowf("Error testing AddFile", "%s", err.Error())
	}

	gs.Require().FileExists(testFile)
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
