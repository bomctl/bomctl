// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/db/db_test.go
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
package db_test

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

type dbSuite struct {
	suite.Suite
	backend   *db.Backend
	documents []*sbom.Document
}

func (dbs *dbSuite) SetupSuite() {
	rdr := reader.New()

	sboms, err := os.ReadDir("testdata")
	if err != nil {
		dbs.T().Fatalf("%v", err)
	}

	for sbomIdx := range sboms {
		doc, err := rdr.ParseFile(filepath.Join("testdata", sboms[sbomIdx].Name()))
		if err != nil {
			dbs.T().Fatalf("%v", err)
		}

		dbs.documents = append(dbs.documents, doc)
	}

	dbs.backend = db.NewBackend()
	dbs.backend.Options.DatabaseFile = db.DatabaseFile

	if err := dbs.backend.InitClient(); err != nil {
		dbs.T().Fatalf("%v", err)
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

func (dbs *dbSuite) TestAddDocument() {
	for _, document := range dbs.documents {
		if err := dbs.backend.AddDocument(document); err != nil {
			dbs.Fail("failed storing document", "id", document.Metadata.Id)
		}
	}
}

func (dbs *dbSuite) TestGetDocumentByID() {
	for _, document := range dbs.documents {
		retrieved, err := dbs.backend.GetDocumentByID(document.Metadata.Id)
		if err != nil {
			dbs.Fail("failed retrieving document", "id", document.Metadata.Id)
		}

		expectedEdges := consolidateEdges(document.NodeList.Edges)
		actualEdges := consolidateEdges(retrieved.NodeList.Edges)

		dbs.Require().Equal(document.Metadata.Id, retrieved.Metadata.Id)
		dbs.Require().Len(retrieved.NodeList.Nodes, len(document.NodeList.Nodes))
		dbs.Require().Equal(expectedEdges, actualEdges)
		dbs.Require().Equal(document.NodeList.RootElements, retrieved.NodeList.RootElements)
	}
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

	for _, e := range edges {
		key := struct {
			fromID   string
			edgeType string
		}{e.From, e.Type.String()}

		edgeMap[key] = append(edgeMap[key], e.To...)
	}

	for typedEdge, toIDs := range edgeMap {
		if len(toIDs) > 0 {
			edgeType := sbom.Edge_Type_value[typedEdge.edgeType]

			consolidated = append(consolidated, &sbom.Edge{
				Type: sbom.Edge_Type(edgeType),
				From: typedEdge.fromID,
				To:   toIDs,
			})
		}
	}

	slices.SortStableFunc(consolidated, func(a, b *sbom.Edge) int { return cmp.Compare(a.From, b.From) })

	return consolidated
}
