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
	"path"

	"github.com/protobom/protobom/pkg/formats"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

func (gs *gitSuite) TestPush() {
	for _, document := range gs.docs {
		rawURL := "git+https://git@github.com:12345/test/repo.git@main#sbom.cdx.json"

		opts := &options.PushOptions{
			Options: gs.opts,
			Format:  formats.CDX15JSON,
			UseTree: false,
		}

		err := gs.gc.Push(document.GetMetadata().GetId(), rawURL, opts)
		if err != nil {
			gs.T().Logf("Error testing Push: %s", err.Error())
		}
	}
}

func (gs *gitSuite) TestAddFile() {
	pOptions := options.PushOptions{
		Options: gs.opts,
		Format:  formats.CDX15JSON,
	}

	parsedURL := &url.ParsedURL{
		Scheme:   "https",
		Username: "git",
		Hostname: "github.com",
		Path:     "test/repo.git",
		GitRef:   "main",
		Fragment: "test/file.sbom",
	}

	err := git.AddFile(gs.repo, path.Join(gs.tempDir, "test", "file.sbom"), &pOptions, gs.doc, parsedURL)
	if err != nil {
		gs.T().Logf("Error testing addFile: %s", err.Error())
	}

	gs.Assert().FileExists(path.Join(gs.tempDir, "test", "file.sbom"))
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
