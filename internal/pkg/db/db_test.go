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
	backend, err := db.NewBackend(db.WithDatabaseFile(":memory:"))
	if err != nil {
		dbs.T().Fatalf("%v", err)
	}

	dbs.backend = backend

	sboms, err := os.ReadDir("testdata")
	if err != nil {
		dbs.T().Fatalf("%v", err)
	}

	for idx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join("testdata", sboms[idx].Name()))
		if err != nil {
			dbs.T().Fatalf("%v", err)
		}

		doc, err := dbs.backend.AddDocument(sbomData)
		if err != nil {
			dbs.FailNow("failed storing document", "err", err)
		}

		name := strings.Split(sboms[idx].Name(), ".")[1]
		dbs.Require().NoError(
			backend.SetUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation, name),
			"failed to set alias", "err", err,
		)

		dbs.Require().NoError(
			backend.AddAnnotations(doc.GetMetadata().GetId(), db.TagAnnotation, documentTags[idx]...),
			"failed to add tags", "err", err,
		)

		dbs.documents = append(dbs.documents, doc)
	}
}

func (dbs *dbSuite) TearDownSuite() {
	dbs.backend.CloseClient()

	if _, err := os.Stat(":memory:"); err == nil {
		if err := os.Remove(":memory:"); err != nil {
			dbs.T().Logf("Error removing database file %s", ":memory:")
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
			expected: dbs.documents[0:2],
		},
		{
			name:     "Normal (another tag, 1 doc)",
			tags:     []string{"tag3"},
			expected: []*sbom.Document{dbs.documents[1]},
		},
		{
			name:     "Normal (multiple tags)",
			tags:     []string{"tag1", "tag2", "tag3"},
			expected: dbs.documents[0:2],
		},
		{
			name:     "Unknown tag",
			tags:     []string{"unknown_tag"},
			expected: []*sbom.Document{},
		},
		{
			name:     "No tags",
			tags:     []string{},
			expected: dbs.documents[0:2],
		},
	} {
		dbs.Run(data.name, func() {
			filteredDocs, err := dbs.backend.FilterDocumentsByTag(docs, data.tags...)
			dbs.Require().NoError(err)
			dbs.Require().Len(filteredDocs, len(data.expected))

			for idx := range data.expected {
				dbs.Equal(filteredDocs[idx].GetMetadata().GetId(), data.expected[idx].GetMetadata().GetId())
			}
		})
	}
}

func (dbs *dbSuite) TestSetAlias() {
	docs, err := dbs.backend.GetDocumentsByID()
	dbs.Require().NoError(err)

	for _, data := range []struct {
		name                  string
		alias                 string
		errorMsg              string
		force                 bool
		removeAliasBeforeTest bool
	}{
		{
			name:                  "Normal",
			alias:                 "cdx",
			errorMsg:              "",
			force:                 false,
			removeAliasBeforeTest: true,
		},
		{
			name:                  "Duplicate alias",
			alias:                 "spdx",
			errorMsg:              "failed to set alias: alias already exists",
			force:                 false,
			removeAliasBeforeTest: true,
		},
		{
			name:                  "Duplicate alias (force)",
			alias:                 "spdx",
			errorMsg:              "failed to set alias: alias already exists",
			force:                 true,
			removeAliasBeforeTest: true,
		},
		{
			name:                  "Existing alias",
			alias:                 "cdx2",
			errorMsg:              "the document already has an alias",
			force:                 false,
			removeAliasBeforeTest: false,
		},
		{
			name:                  "Existing alias (force)",
			alias:                 "cdx2",
			errorMsg:              "",
			force:                 true,
			removeAliasBeforeTest: false,
		},
	} {
		dbs.Run(data.name, func() {
			if data.removeAliasBeforeTest {
				err := dbs.backend.RemoveAnnotations(docs[0].GetMetadata().GetId(), db.AliasAnnotation, "cdx")
				dbs.Require().NoError(err)
			}

			err = dbs.backend.SetAlias(docs[0].GetMetadata().GetId(), data.alias, data.force)
			if data.errorMsg == "" {
				dbs.Require().NoError(err)
				docAlias, err := dbs.backend.GetDocumentUniqueAnnotation(docs[0].GetMetadata().GetId(), db.AliasAnnotation)
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
