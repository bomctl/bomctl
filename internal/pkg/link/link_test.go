// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/link/link_test.go
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

package link_test

import (
	"testing"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/link"
	"github.com/bomctl/bomctl/internal/pkg/logger"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
	"github.com/bomctl/bomctl/internal/testutil"
)

type linkSuite struct {
	suite.Suite
	*db.Backend
	documents    []*sbom.Document
	documentInfo []testutil.DocumentInfo
}

var (
	cdxDocTarget = options.LinkTarget{
		ID:    "urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79",
		Alias: "cdx",
		Type:  options.LinkTargetTypeDocument,
	}

	spdxDocTarget = options.LinkTarget{
		ID:    "https://spdx.org/spdxdocs/apko/#DOCUMENT",
		Alias: "spdx",
		Type:  options.LinkTargetTypeDocument,
	}

	nodeTarget = options.LinkTarget{
		ID:   "Package-libbrotlicommon1-1.0.9-r3",
		Type: options.LinkTargetTypeNode,
	}
)

func (ls *linkSuite) SetupSubTest() {
	var err error

	ls.Backend, err = testutil.NewTestBackend()
	ls.Require().NoError(err, "failed database backend creation")

	ls.documentInfo, err = testutil.AddTestDocuments(ls.Backend)
	ls.Require().NoError(err, "failed database backend setup")

	for _, docInfo := range ls.documentInfo {
		ls.documents = append(ls.documents, docInfo.Document)
	}
}

func (ls *linkSuite) TearDownSubTest() {
	ls.Backend.CloseClient()
}

func (ls *linkSuite) TestAddLink() {
	opts := options.New().WithLogger(logger.New("link_add_test"))

	subtests := []struct {
		name     string
		from, to options.LinkTarget
	}{
		{name: "document-document", from: spdxDocTarget, to: cdxDocTarget},
		{name: "node-document", from: nodeTarget, to: cdxDocTarget},
	}

	for _, subtest := range subtests {
		linkOpts := &options.LinkOptions{
			Options: opts,
			Links:   []options.Link{{From: subtest.from, To: []options.LinkTarget{subtest.to}}},
		}

		ls.Run(subtest.name, func() {
			ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))

			annotations := ent.Annotations{}

			var err error

			switch subtest.from.Type {
			case options.LinkTargetTypeDocument:
				annotations, err = ls.Backend.GetDocumentAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
			case options.LinkTargetTypeNode:
				annotations, err = ls.Backend.GetNodeAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
			}

			ls.Require().NotEmpty(annotations)
			lastAnnotation := annotations[len(annotations)-1]

			ls.Require().Equal(lastAnnotation.Value, subtest.to.ID)
		})
	}
}

func (ls *linkSuite) TestClearLinks() {
	opts := options.New().WithLogger(logger.New("link_clear_test"))

	subtests := []struct {
		name string
		from []options.LinkTarget
	}{
		{name: "document", from: []options.LinkTarget{spdxDocTarget}},
		{name: "node", from: []options.LinkTarget{nodeTarget}},
		{name: "document-and-node", from: []options.LinkTarget{spdxDocTarget, nodeTarget}},
	}

	for _, subtest := range subtests {
		ls.Run(subtest.name, func() {
			for _, from := range subtest.from {
				linkOpts := &options.LinkOptions{
					Options: opts,
					Links:   []options.Link{{From: from, To: []options.LinkTarget{cdxDocTarget}}},
				}

				// Add links to verify clearing functionality.
				ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))

				switch from.Type {
				case options.LinkTargetTypeDocument:
					annotations, err := ls.Backend.GetDocumentAnnotations(from.ID, db.LinkToAnnotation)
					ls.Require().NoError(err)
					ls.NotEmpty(annotations)
				case options.LinkTargetTypeNode:
					annotations, err := ls.Backend.GetNodeAnnotations(from.ID, db.LinkToAnnotation)
					ls.Require().NoError(err)
					ls.NotEmpty(annotations)
				}
			}

			ls.Require().NoError(link.ClearLinks(ls.Backend, &options.LinkOptions{
				Options: opts,
				Links: sliceutil.Extract(subtest.from, func(lt options.LinkTarget) options.Link {
					return options.Link{From: lt, To: []options.LinkTarget{cdxDocTarget}}
				}),
			}))

			for _, from := range subtest.from {
				switch from.Type {
				case options.LinkTargetTypeDocument:
					annotations, err := ls.Backend.GetDocumentAnnotations(from.ID, db.LinkToAnnotation)
					ls.Require().NoError(err)
					ls.Empty(annotations)
				case options.LinkTargetTypeNode:
					annotations, err := ls.Backend.GetNodeAnnotations(from.ID, db.LinkToAnnotation)
					ls.Require().NoError(err)
					ls.Empty(annotations)
				}
			}
		})
	}
}

func (ls *linkSuite) TestListLinks() {
	opts := options.New().WithLogger(logger.New("link_list_test"))
	subtests := []struct {
		name string
		from options.LinkTarget
	}{
		{name: "document", from: spdxDocTarget},
		{name: "node", from: nodeTarget},
	}

	for _, subtest := range subtests {
		ls.Run(subtest.name, func() {
			linkOpts := &options.LinkOptions{
				Options: opts,
				Links:   []options.Link{{From: subtest.from}},
			}

			_, err := link.ListLinks(ls.Backend, linkOpts)
			ls.Require().NoError(err)
			ls.Empty(linkOpts.Links[0].To)

			linkOpts.Links = []options.Link{{From: subtest.from, To: []options.LinkTarget{cdxDocTarget}}}

			// Add link to verify list.
			ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))

			_, err = link.ListLinks(ls.Backend, linkOpts)
			ls.Require().NoError(err)
			ls.NotEmpty(linkOpts.Links[0].To)

			linkOpts.Links = []options.Link{{From: cdxDocTarget}}

			incoming, err := link.ListLinks(ls.Backend, linkOpts)
			ls.Require().NoError(err)
			ls.Empty(linkOpts.Links[0].To)

			ls.ElementsMatch(incoming, []options.LinkTarget{subtest.from})
		})
	}
}

func (ls *linkSuite) TestRemoveLink() {
	opts := options.New().WithLogger(logger.New("link_remove_test"))

	subtests := []struct {
		name     string
		from, to options.LinkTarget
	}{
		{name: "document-document", from: spdxDocTarget, to: cdxDocTarget},
		{name: "node-document", from: nodeTarget, to: cdxDocTarget},
	}

	for _, subtest := range subtests {
		ls.Run(subtest.name, func() {
			linkOpts := &options.LinkOptions{
				Options: opts,
				Links:   []options.Link{{From: subtest.from, To: []options.LinkTarget{subtest.to}}},
			}

			// Add link to verify removal.
			ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))

			switch subtest.from.Type {
			case options.LinkTargetTypeDocument:
				annotations, err := ls.Backend.GetDocumentAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.NotEmpty(annotations)

				ls.Require().NoError(link.RemoveLink(ls.Backend, linkOpts))

				annotations, err = ls.Backend.GetDocumentAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.Empty(annotations)
			case options.LinkTargetTypeNode:
				annotations, err := ls.Backend.GetNodeAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.NotEmpty(annotations)

				ls.Require().NoError(link.RemoveLink(ls.Backend, linkOpts))

				annotations, err = ls.Backend.GetNodeAnnotations(subtest.from.ID, db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.Empty(annotations)
			}
		})
	}
}

func TestLinkSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(linkSuite))
}
