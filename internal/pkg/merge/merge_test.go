// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/merge/merge_test.go
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
package merge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/merge"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

var TestDataDir = filepath.Join("..", "db", "testdata")

type mergeSuite struct {
	suite.Suite
	opts    *options.Options
	backend *db.Backend
	docs    []*sbom.Document
}

func (ms *mergeSuite) SetupSuite() {
	sboms, err := os.ReadDir(TestDataDir)
	if err != nil {
		ms.T().Fatalf("%v", err)
	}

	rdr := reader.New()
	for sbomIdx := range sboms {
		doc, err := rdr.ParseFile(filepath.Join(TestDataDir, sboms[sbomIdx].Name()))
		if err != nil {
			ms.T().Fatalf("%v", err)
		}

		ms.docs = append(ms.docs, doc)
	}

	ms.opts = options.New().
		WithCacheDir(viper.GetString("cache_dir"))

	ms.backend, err = db.NewBackend(
		db.WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)),
	)
	if err != nil {
		ms.T().Fatalf("%v", err)
	}

	for _, document := range ms.docs {
		err := ms.backend.AddDocument(document)
		if err != nil {
			ms.Fail("failed retrieving document", "id", document.GetMetadata().GetId())
		}
	}

	ms.opts = ms.opts.WithContext(context.WithValue(context.Background(), db.BackendKey{}, ms.backend))
}

func (ms *mergeSuite) TearDownSuite() {
	ms.backend.CloseClient()

	if _, err := os.Stat(db.DatabaseFile); err == nil {
		if err := os.Remove(db.DatabaseFile); err != nil {
			ms.T().Logf("Error removing database file %s", db.DatabaseFile)
		}
	}
}

func (ms *mergeSuite) TestMerge() {
	opts := &options.MergeOptions{
		Options: ms.opts,
	}

	docID, err := merge.Merge([]string{ms.docs[0].GetMetadata().GetId(), ms.docs[1].GetMetadata().GetId()}, opts)

	ms.Require().NoError(err)

	mergedDoc, err := ms.backend.GetDocumentByID(docID)
	if err != nil {
		ms.Fail("Failed to get merged document from DB")
	}

	if ms.docs[0].GetMetadata().GetName() != "" {
		ms.Equal(ms.docs[0].GetMetadata().GetName(), mergedDoc.GetMetadata().GetName())
	} else if ms.docs[0].GetMetadata().GetName() == "" && ms.docs[1].GetMetadata().GetName() != "" {
		ms.Equal(ms.docs[1].GetMetadata().GetName(), mergedDoc.GetMetadata().GetName())
	}

	expectedToolLength := len(ms.docs[0].GetMetadata().GetTools()) + len(ms.docs[1].GetMetadata().GetTools())

	ms.Len(mergedDoc.GetMetadata().GetTools(), expectedToolLength)
}

func TestMergeSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(mergeSuite))
}
