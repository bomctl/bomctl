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
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

type dbSuite struct {
	suite.Suite
	backend   *db.Backend
	documents []*sbom.Document
}

var documentTags = [][]string{{"tag1", "tag2"}, {"tag2", "tag3"}}

func (dbs *dbSuite) SetupSuite() {
	backend, err := db.NewBackend(db.WithDatabaseFile(db.DatabaseFile))
	if err != nil {
		dbs.T().Fatalf("%v", err)
	}

	dbs.backend = backend

	sboms, err := os.ReadDir("testdata")
	if err != nil {
		dbs.T().Fatalf("%v", err)
	}

	for sbomIdx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join("testdata", sboms[sbomIdx].Name()))
		if err != nil {
			dbs.T().Fatalf("%v", err)
		}

		doc, err := dbs.backend.AddDocument(sbomData)
		if err != nil {
			dbs.FailNow("failed storing document", "err", err)
		}

		name := strings.Split(sboms[sbomIdx].Name(), ".")[1]
		if err := backend.SetUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation, name); err != nil {
			dbs.FailNow("failed to set alias", "err", err)
		}

		if err := backend.AddAnnotations(doc.GetMetadata().GetId(),
			db.TagAnnotation, documentTags[sbomIdx]...); err != nil {
			dbs.FailNow("failed to add tags", "err", err)
		}

		dbs.documents = append(dbs.documents, doc)
	}
}

func (dbs *dbSuite) TearDownSuite() {
	dbs.backend.CloseClient()

	if _, err := os.Stat(db.DatabaseFile); err == nil {
		if err := os.Remove(db.DatabaseFile); err != nil {
			dbs.T().Logf("Error removing database file %s", db.DatabaseFile)
		}
	}
}

func (dbs *dbSuite) TestGetDocumentByID() {
	for _, document := range dbs.documents {
		retrieved, err := dbs.backend.GetDocumentByID(document.GetMetadata().GetId())
		if err != nil {
			dbs.Fail("failed retrieving document", "id", document.GetMetadata().GetId())
		}

		expectedEdges := consolidateEdges(document.GetNodeList().GetEdges())
		actualEdges := consolidateEdges(retrieved.GetNodeList().GetEdges())

		dbs.Require().Equal(document.GetMetadata().GetId(), retrieved.GetMetadata().GetId())
		dbs.Require().Len(retrieved.GetNodeList().GetNodes(), len(document.GetNodeList().GetNodes()))
		dbs.Require().Equal(expectedEdges, actualEdges)
		dbs.Require().Equal(document.GetNodeList().GetRootElements(), retrieved.GetNodeList().GetRootElements())
	}
}

func (dbs *dbSuite) TestGetDocumentByIDOrAlias() {
	cdxDoc, err := dbs.backend.GetDocumentByIDOrAlias("cdx")
	if err != nil {
		dbs.Fail("failed retrieving document", "alias", "cdx")
	}

	spdxDoc, err := dbs.backend.GetDocumentByIDOrAlias("spdx")
	if err != nil {
		dbs.Fail("failed retrieving document", "alias", "spdx")
	}

	dbs.Require().Equal(cdxDoc.GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(spdxDoc.GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())
}

func (dbs *dbSuite) TestGetDocumentsByIDOrAlias() {
	docs, err := dbs.backend.GetDocumentsByIDOrAlias("cdx", "spdx")
	if err != nil {
		dbs.Fail("failed retrieving document", "aliases", "cdx, spdx")
	}

	dbs.Require().Equal(docs[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(docs[1].GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())
}

func (dbs *dbSuite) TestGetDocumentTags() {
	tags, err := dbs.backend.GetDocumentTags(dbs.documents[0].GetMetadata().GetId())
	dbs.Require().NoError(err)
	dbs.Require().EqualValues([]string{"tag1", "tag2"}, tags)
}

func (dbs *dbSuite) TestFilterDocumentsByTag() {
	docs, err := dbs.backend.GetDocumentsByID()
	dbs.Require().NoError(err)

	// Normal (1 tag, 1 doc)
	filteredDocs1, err := dbs.backend.FilterDocumentsByTag(docs, "tag1")
	dbs.Require().NoError(err)
	dbs.Require().Len(filteredDocs1, 1)
	dbs.Require().Equal(filteredDocs1[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())

	// Normal (1 tag, 2 docs)
	filteredDocs2, err := dbs.backend.FilterDocumentsByTag(docs, "tag2")
	dbs.Require().NoError(err)
	dbs.Require().Len(filteredDocs2, 2)
	dbs.Require().Equal(filteredDocs2[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(filteredDocs2[1].GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())

	// Normal (another tag, 1 doc)
	filteredDocs3, err := dbs.backend.FilterDocumentsByTag(docs, "tag3")
	dbs.Require().NoError(err)
	dbs.Require().Len(filteredDocs3, 1)
	dbs.Require().Equal(filteredDocs3[0].GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())

	// Normal (multiple tags)
	filteredDocs4, err := dbs.backend.FilterDocumentsByTag(docs, "tag1", "tag2", "tag3")
	dbs.Require().NoError(err)
	dbs.Require().Len(filteredDocs4, 2)
	dbs.Require().Equal(filteredDocs4[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(filteredDocs4[1].GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())

	// Unknown tag
	filteredDocs5, err := dbs.backend.FilterDocumentsByTag(docs, "unknown_tag")
	dbs.Require().NoError(err)
	dbs.Require().Empty(filteredDocs5)

	// No tags
	filteredDocs6, err := dbs.backend.FilterDocumentsByTag(docs)
	dbs.Require().NoError(err)
	dbs.Require().Len(filteredDocs6, 2)
	dbs.Require().Equal(filteredDocs6[0].GetMetadata().GetId(), dbs.documents[0].GetMetadata().GetId())
	dbs.Require().Equal(filteredDocs6[1].GetMetadata().GetId(), dbs.documents[1].GetMetadata().GetId())
}

func (dbs *dbSuite) TestSetAlias() {
	docs, err := dbs.backend.GetDocumentsByID()
	dbs.Require().NoError(err)

	err = dbs.backend.RemoveAnnotations(docs[0].GetMetadata().GetId(), db.AliasAnnotation, "cdx")
	dbs.Require().NoError(err)

	// Error: Duplicate alias
	err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), "spdx", false)
	dbs.Require().EqualError(err, "failed to set alias: alias already exists")

	// Error: Duplicate alias
	err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), "spdx", true)
	dbs.Require().EqualError(err, "failed to set alias: alias already exists")

	// Set Alias (Normal)
	err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), "cdx", false)
	dbs.Require().NoError(err)
	docAlias, err := dbs.backend.GetDocumentUniqueAnnotation(docs[0].GetMetadata().GetId(), db.AliasAnnotation)
	dbs.Require().NoError(err)
	dbs.Require().Equal("cdx", docAlias)

	// Error: Document already has an alias
	err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), "cdx2", false)
	dbs.Require().EqualError(err, "the document already has an alias")

	// Set Alias (Force)
	err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), "cdx2", true)
	dbs.Require().NoError(err)
	docAlias, err = dbs.backend.GetDocumentUniqueAnnotation(docs[0].GetMetadata().GetId(), db.AliasAnnotation)
	dbs.Require().NoError(err)
	dbs.Require().Equal("cdx2", docAlias)
}

func TestStoreSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(dbSuite))
}

func consolidateEdges(edges []*sbom.Edge) []*sbom.Edge {
	consolidated := []*sbom.Edge{}

	// Mapping of from ID and edge type to slice of to IDs.
	edgeMap := make(map[struct {
		fromID   string
		edgeType string
	}][]string)

	for _, edge := range edges {
		key := struct {
			fromID   string
			edgeType string
		}{edge.GetFrom(), edge.GetType().String()}

		edgeMap[key] = append(edgeMap[key], edge.GetTo()...)
	}

	for typedEdge, toIDs := range edgeMap {
		slices.Sort(toIDs)

		if len(toIDs) > 0 {
			slices.Sort(toIDs)

			edgeType := sbom.Edge_Type_value[typedEdge.edgeType]
			consolidated = append(consolidated, &sbom.Edge{
				Type: sbom.Edge_Type(edgeType),
				From: typedEdge.fromID,
				To:   toIDs,
			})
		}
	}

	slices.SortStableFunc(consolidated, func(a, b *sbom.Edge) int { return cmp.Compare(a.GetFrom(), b.GetFrom()) })

	return consolidated
}
