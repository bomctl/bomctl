// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/db/db_test.go
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

package db_test

import (
	"fmt"
	"testing"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/testutil"
)

type dbSuite struct {
	suite.Suite
	*db.Backend
	documents    []*sbom.Document
	documentInfo []testutil.DocumentInfo
}

func (dbs *dbSuite) SetupTest() {
	var err error

	dbs.Backend, err = testutil.NewTestBackend()
	dbs.Require().NoError(err, "failed database backend creation")

	dbs.documentInfo, err = testutil.AddTestDocuments(dbs.Backend)
	dbs.Require().NoError(err, "failed database backend setup")

	for _, docInfo := range dbs.documentInfo {
		dbs.documents = append(dbs.documents, docInfo.Document)
	}
}

func (dbs *dbSuite) TearDownTest() {
	dbs.Backend.CloseClient()
	dbs.documents = nil
	dbs.documentInfo = nil
}

func (dbs *dbSuite) TestBackend_AddDocumentRevision() {
	for _, data := range []struct {
		prep     func()
		name     string
		baseID   string
		alias    string
		errorMsg string
	}{
		{
			name:   "existing alias on base doc",
			prep:   func() {},
			baseID: "8daeb29e-8655-fae1-b792-78b998823fc6",
			alias:  "cdx",
		},
		{
			name: "no alias on base doc",
			prep: func() {
				err := dbs.Backend.RemoveDocumentAnnotations(
					dbs.documentInfo[0].Document.GetMetadata().GetId(),
					db.AliasAnnotation)
				dbs.Require().NoError(err)
			},
			baseID: "8daeb29e-8655-fae1-b792-78b998823fc6",
			alias:  "",
		},
	} {
		dbs.Run(data.name, func() {
			docContent := dbs.documentInfo[1].Content
			baseDoc := dbs.documentInfo[0].Document

			data.prep()

			err := dbs.Backend.ClearDocumentAnnotations(dbs.documentInfo[1].Document.GetMetadata().GetId())
			dbs.Require().NoError(err)

			newDoc, err := dbs.Backend.AddDocument(docContent, db.WithRevisedDocumentAnnotations(baseDoc))

			if data.errorMsg == "" {
				dbs.Require().NoError(err)

				newID := newDoc.GetMetadata().GetId()
				baseDocID := dbs.documentInfo[0].Document.GetMetadata().GetId()

				newUUID, err := ent.GenerateUUID(newDoc)
				dbs.Require().NoError(err)

				baseID, err := dbs.Backend.GetDocumentUniqueAnnotation(newID, db.BaseDocumentAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal(data.baseID, baseID)

				alias, err := dbs.Backend.GetDocumentUniqueAnnotation(newID, db.AliasAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal(data.alias, alias)

				latest, err := dbs.Backend.GetDocumentAnnotations(newID, db.LatestRevisionAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal("true", latest[0].Value)

				latest, err = dbs.Backend.GetDocumentAnnotations(baseDocID, db.LatestRevisionAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal(ent.Annotations{}, latest)

				revision, err := dbs.Backend.GetDocumentAnnotations(baseDocID, db.RevisedDocumentAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal(newUUID.String(), revision[0].Value)

				alias, err = dbs.Backend.GetDocumentUniqueAnnotation(baseDocID, db.AliasAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal("", alias)

				srcData, err := dbs.Backend.GetDocumentUniqueAnnotation(newID, db.SourceDataAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal("", srcData)
			} else {
				dbs.Require().EqualError(err, data.errorMsg)
			}
		})
	}
}

func (dbs *dbSuite) TestBackend_FilterDocumentsByTag() {
	for _, data := range []struct {
		name     string
		tags     []string
		expected []*sbom.Document
	}{
		{
			name:     "Normal (1 tag, 1 doc)",
			tags:     []string{"tag1"},
			expected: []*sbom.Document{dbs.documents[0]},
		},
		{
			name:     "Normal (1 tag, 2 docs)",
			tags:     []string{"tag2"},
			expected: dbs.documents[:2],
		},
		{
			name:     "Normal (another tag, 1 doc)",
			tags:     []string{"tag5"},
			expected: []*sbom.Document{dbs.documents[3]},
		},
		{
			name:     "Normal (multiple tags)",
			tags:     []string{"tag1", "tag2", "tag3"},
			expected: dbs.documents[0:3],
		},
		{
			name:     "Unknown tag",
			tags:     []string{"unknown_tag"},
			expected: []*sbom.Document{},
		},
		{
			name:     "No tags",
			tags:     []string{},
			expected: dbs.documents,
		},
	} {
		dbs.Run(data.name, func() {
			filteredDocs, err := dbs.Backend.FilterDocumentsByTag(dbs.documents, data.tags...)
			dbs.Require().NoError(err)
			dbs.Require().Len(
				filteredDocs,
				len(data.expected),
				fmt.Sprintf("expected length: %d does not match actual: %d", len(filteredDocs), len(data.expected)),
			)

			for idx := range data.expected {
				dbs.Equal(filteredDocs[idx].GetMetadata().GetId(), data.expected[idx].GetMetadata().GetId())
			}
		})
	}
}

func (dbs *dbSuite) TestBackend_GetDocumentByID() {
	for _, document := range dbs.documents {
		retrieved, err := dbs.Backend.GetDocumentByID(document.GetMetadata().GetId())
		dbs.Require().NoError(err, "failed retrieving document", "id", document.GetMetadata().GetId())

		dbs.Require().True(retrieved.GetNodeList().Equal(document.GetNodeList()))
		dbs.Require().Equal(document.GetMetadata().GetId(), retrieved.GetMetadata().GetId())
	}
}

func (dbs *dbSuite) TestBackend_GetDocumentByIDOrAlias() {
	cdxDoc, err := dbs.Backend.GetDocumentByIDOrAlias("cdx")
	dbs.Require().NoError(err, "failed retrieving document with alias: cdx")

	spdxDoc, err := dbs.Backend.GetDocumentByIDOrAlias("spdx")
	dbs.Require().NoError(err, "failed retrieving document with alias: spdx")

	dbs.Require().Equal(dbs.documents[0].GetMetadata().GetId(), cdxDoc.GetMetadata().GetId())
	dbs.Require().Equal(dbs.documents[2].GetMetadata().GetId(), spdxDoc.GetMetadata().GetId())
}

func (dbs *dbSuite) TestBackend_GetDocumentsByIDOrAlias() {
	docs, err := dbs.Backend.GetDocumentsByIDOrAlias("cdx", "spdx")
	dbs.Require().NoError(err, "failed retrieving documents with aliases: cdx, spdx")

	dbs.Require().Equal(docs[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(docs[1].GetMetadata().GetId(), dbs.documents[2].GetMetadata().GetId())
}

func (dbs *dbSuite) TestBackend_GetDocumentTags() {
	tags, err := dbs.Backend.GetDocumentTags(dbs.documents[0].GetMetadata().GetId())
	dbs.Require().NoError(err)
	dbs.Require().EqualValues([]string{"tag1", "tag2"}, tags)
}

func (dbs *dbSuite) TestBackend_GetLatestDocument() {
	tempbe, err := testutil.NewTestBackend()
	dbs.Require().NoError(err)

	boms, err := createTestDocumentSet(dbs.documentInfo, tempbe)
	dbs.Require().NoError(err)

	for _, bom := range boms {
		rootDoc, err := tempbe.GetLatestDocument(bom)
		dbs.Require().NoError(err)
		dbs.Require().Equal(rootDoc.GetMetadata().GetId(), boms[3].GetMetadata().GetId())
	}
}

func (dbs *dbSuite) TestBackend_GetRootDocument() {
	tempbe, err := testutil.NewTestBackend()
	dbs.Require().NoError(err)

	boms, err := createTestDocumentSet(dbs.documentInfo, tempbe)
	dbs.Require().NoError(err)

	for _, bom := range boms {
		rootDoc, err := tempbe.GetRootDocument(bom)
		dbs.Require().NoError(err)
		dbs.Require().Equal(rootDoc.GetMetadata().GetId(), boms[0].GetMetadata().GetId())
	}
}

func (dbs *dbSuite) TestBackend_SetAlias() {
	id := dbs.documents[0].GetMetadata().GetId()

	for _, data := range []struct {
		name      string
		alias     string
		errorMsg  string
		doc0Alias string
		force     bool
	}{
		{
			name:      "Normal",
			alias:     "cdx",
			errorMsg:  "",
			doc0Alias: "",
			force:     false,
		},
		{
			name:      "Duplicate alias",
			alias:     "spdx",
			errorMsg:  "failed to set alias: alias already exists",
			doc0Alias: "",
			force:     false,
		},
		{
			name:      "Duplicate alias (force)",
			alias:     "spdx",
			errorMsg:  "failed to set alias: alias already exists",
			doc0Alias: "",
			force:     true,
		},
		{
			name:      "Existing alias",
			alias:     "cdx16",
			errorMsg:  "the document already has an alias",
			doc0Alias: "cdx",
			force:     false,
		},
		{
			name:      "Existing alias (force)",
			alias:     "cdx16",
			errorMsg:  "",
			doc0Alias: "cdx",
			force:     true,
		},
	} {
		dbs.Run(data.name, func() {
			err := dbs.Backend.RemoveDocumentAnnotations(id, db.AliasAnnotation, "cdx")
			dbs.Require().NoError(err)

			if data.doc0Alias != "" {
				dbs.Require().NoError(
					dbs.Backend.SetDocumentUniqueAnnotation(id, db.AliasAnnotation, data.doc0Alias),
					"failed to set alias", "err", err,
				)
			}

			err = dbs.Backend.SetAlias(id, data.alias, data.force)
			if data.errorMsg == "" {
				dbs.Require().NoError(err)
				docAlias, err := dbs.Backend.GetDocumentUniqueAnnotation(id, db.AliasAnnotation)
				dbs.Require().NoError(err)
				dbs.Require().Equal(data.alias, docAlias)
			} else {
				dbs.Require().EqualError(err, data.errorMsg)
			}
		})
	}
}

func TestDBSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(dbSuite))
}

func createTestDocumentSet(sboms []testutil.DocumentInfo, backend *db.Backend) ([]*sbom.Document, error) {
	rootDoc, err := backend.AddDocument(sboms[0].Content, db.WithSourceDocumentAnnotations(sboms[0].Content))
	if err != nil {
		return nil, fmt.Errorf("creating test document set: %w", err)
	}

	docSlice := []*sbom.Document{rootDoc}

	for i, bom := range sboms[1:] {
		doc, err := backend.AddDocument(bom.Content, db.WithRevisedDocumentAnnotations(sboms[i].Document))
		if err != nil {
			return nil, fmt.Errorf("creating test document set: %w", err)
		}

		docSlice = append(docSlice, doc)
	}

	return docSlice, nil
}
