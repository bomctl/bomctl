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

	"github.com/protobom/protobom/pkg/formats"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (gs *gitSuite) TestPush() {
	for _, document := range gs.docs {
		rawURL := "git+https://git@github.com:12345/test/repo.git@main#sbom.cdx.json"

		opts := &options.PushOptions{
			Options: gs.opts,
			Format:  formats.CDX15JSON,
			UseTree: false,
		}

		err := gs.gc.Push(document.Metadata.Id, rawURL, opts)
		if err != nil {
			gs.T().Logf("Error testing Push: %s", err.Error())
		}
	}
}

func (gs *gitSuite) TestAddFile() {
	opts := options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
	}

	testFile := filepath.Join(gs.tmpDir, "test", "file.sbom")

	if err := os.MkdirAll(filepath.Dir(testFile), fs.ModePerm); err != nil {
		gs.FailNow("failed creating directory")
	}

	file, err := os.Create(testFile)
	if err != nil {
		gs.FailNow("failed creating file")
	}

	if err := gs.gc.AddFile(file, gs.doc, &opts); err != nil {
		gs.FailNowf("Error testing AddFile", "%s", err.Error())
	}

	gs.Require().FileExists(testFile)
}

func (gs *gitSuite) TestGetDocument() {
	for _, document := range gs.docs {
		retrieved, err := git.GetDocument(document.Metadata.Id, gs.opts)
		if err != nil {
			gs.FailNow("failed retrieving document", "id", document.Metadata.Id)
		}

		gs.Require().Equal(document.Metadata.Id, retrieved.Metadata.Id)
		gs.Require().Len(retrieved.NodeList.Nodes, len(document.NodeList.Nodes))
		gs.Require().Equal(document.NodeList.RootElements, retrieved.NodeList.RootElements)
	}
}
