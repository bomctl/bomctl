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
	"github.com/bomctl/bomctl/internal/testutil"
)

type linkSuite struct {
	suite.Suite
	*db.Backend
	documents    []*sbom.Document
	documentInfo []testutil.DocumentInfo
}

func (ls *linkSuite) SetupSuite() {
	var err error

	ls.Backend, err = testutil.NewTestBackend()
	ls.Require().NoError(err, "failed database backend creation")

	ls.documentInfo, err = testutil.AddTestDocuments(ls.Backend)
	ls.Require().NoError(err, "failed database backend setup")

	for _, docInfo := range ls.documentInfo {
		ls.documents = append(ls.documents, docInfo.Document)
	}
}

func (ls *linkSuite) TearDownSuite() {
	ls.Backend.CloseClient()
}

func (ls *linkSuite) TestAddLink() {
	opts := options.New().WithLogger(logger.New("link_add_test"))

	subtests := []struct {
		name        string
		documentIDs []string
		nodeIDs     []string
		toIDs       []string
	}{
		{
			name:        "alias-document",
			documentIDs: []string{"spdx"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:        "document-document",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:    "node-document",
			nodeIDs: []string{"Package-libbrotlicommon1-1.0.9-r3"},
			toIDs:   []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
	}

	for _, subtest := range subtests {
		linkOpts := &options.LinkOptions{
			Options:     opts,
			DocumentIDs: subtest.documentIDs,
			NodeIDs:     subtest.nodeIDs,
			ToIDs:       subtest.toIDs,
		}

		ls.Run(subtest.name, func() {
			ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))

			annotations := ent.Annotations{}

			if len(linkOpts.DocumentIDs) > 0 {
				docID, err := ls.Backend.GetDocumentByIDOrAlias(linkOpts.DocumentIDs[0])
				ls.Require().NoError(err)

				annotations, err = ls.Backend.GetDocumentAnnotations(docID.GetMetadata().GetId())
				ls.Require().NoError(err)
			}

			if len(linkOpts.NodeIDs) > 0 {
				var err error

				annotations, err = ls.Backend.GetNodeAnnotations(linkOpts.NodeIDs[0])
				ls.Require().NoError(err)
			}

			ls.Require().NotEmpty(annotations)
			lastAnnotation := annotations[len(annotations)-1]

			ls.Require().Equal(lastAnnotation.Value, linkOpts.ToIDs[0])
		})
	}
}

func (ls *linkSuite) TestClearLinks() {
	opts := options.New().WithLogger(logger.New("link_clear_test"))

	subtests := []struct {
		name        string
		documentIDs []string
		nodeIDs     []string
	}{
		{
			name:        "alias-document",
			documentIDs: []string{"spdx"},
		},
		{
			name:        "document",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
		},
		{
			name:    "node",
			nodeIDs: []string{"Package-libbrotlicommon1-1.0.9-r3"},
		},
		{
			name:        "document-and-node",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
			nodeIDs:     []string{"Package-libbrotlicommon1-1.0.9-r3"},
		},
	}

	for _, subtest := range subtests {
		linkOpts := &options.LinkOptions{
			Options:     opts,
			DocumentIDs: subtest.documentIDs,
			NodeIDs:     subtest.nodeIDs,
		}

		ls.Run(subtest.name, func() {
			ls.Require().NoError(link.ClearLinks(ls.Backend, linkOpts))

			if len(linkOpts.DocumentIDs) > 0 {
				annotations, err := ls.Backend.GetDocumentAnnotations(subtest.documentIDs[0], db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.Empty(annotations)
			}

			if len(linkOpts.NodeIDs) > 0 {
				annotations, err := ls.Backend.GetNodeAnnotations(subtest.nodeIDs[0], db.LinkToAnnotation)
				ls.Require().NoError(err)
				ls.Empty(annotations)
			}
		})
	}
}

func (ls *linkSuite) TestListLinks() {
	opts := options.New().WithLogger(logger.New("link_list_test"))
	subtests := []struct {
		name        string
		documentIDs []string
		nodeIDs     []string
		toIDs       []string
	}{
		{
			name:        "alias-document",
			documentIDs: []string{"spdx"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:        "document",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:    "node",
			nodeIDs: []string{"Package-libbrotlicommon1-1.0.9-r3"},
			toIDs:   []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:        "document-and-node",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
			nodeIDs:     []string{"Package-libbrotlicommon1-1.0.9-r3"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
	}

	for _, subtest := range subtests {
		linkOpts := &options.LinkOptions{
			Options:     opts,
			DocumentIDs: subtest.documentIDs,
			NodeIDs:     subtest.nodeIDs,
			ToIDs:       subtest.toIDs,
		}

		ls.Run(subtest.name, func() {
			ls.Require().NoError(link.AddLink(ls.Backend, linkOpts))
			ls.Require().NoError(link.ListLinks(ls.Backend, linkOpts))
		})
	}
}

func (ls *linkSuite) TestRemoveLink() {
	opts := options.New().WithLogger(logger.New("link_remove_test"))

	subtests := []struct {
		name        string
		documentIDs []string
		nodeIDs     []string
		toIDs       []string
	}{
		{
			name:        "alias-document",
			documentIDs: []string{"spdx"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:        "document-document",
			documentIDs: []string{"https://spdx.org/spdxdocs/apko/#DOCUMENT"},
			toIDs:       []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
		{
			name:    "node-document",
			nodeIDs: []string{"Package-libbrotlicommon1-1.0.9-r3"},
			toIDs:   []string{"urn:uuid:3e671687-395b-41f5-a30f-a58921a69b79"},
		},
	}

	for _, subtest := range subtests {
		linkOpts := &options.LinkOptions{
			Options:     opts,
			DocumentIDs: subtest.documentIDs,
			NodeIDs:     subtest.nodeIDs,
			ToIDs:       subtest.toIDs,
		}

		ls.Run(subtest.name, func() {
			ls.Require().NoError(link.RemoveLink(ls.Backend, linkOpts))

			names := []string{}

			if len(linkOpts.DocumentIDs) > 0 {
				annotations, err := ls.Backend.GetDocumentAnnotations(subtest.documentIDs[0])
				ls.Require().NoError(err)

				for _, annotation := range annotations {
					names = append(names, annotation.Name)
				}
			}

			if len(linkOpts.NodeIDs) > 0 {
				annotations, err := ls.Backend.GetNodeAnnotations(subtest.nodeIDs[0])
				ls.Require().NoError(err)

				for _, annotation := range annotations {
					names = append(names, annotation.Name)
				}
			}

			ls.Require().NotContains(names, "bomctl_annotation_link_to")
		})
	}
}

func TestLinkSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(linkSuite))
}