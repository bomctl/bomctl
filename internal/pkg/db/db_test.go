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
package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bom-squad/protobom/pkg/reader"
	protobom "github.com/bom-squad/protobom/pkg/sbom"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	sbomReader   = reader.New()
	errConvertPB = errors.New("bomctl-test")
)

type mockDocument struct {
	protobom.Document
	mock.Mock
	shouldError bool
}

func (md *mockDocument) ToORM(ctx context.Context) (protobom.DocumentORM, error) {
	var documentORM protobom.DocumentORM

	if md.shouldError {
		return protobom.DocumentORM{}, fmt.Errorf("failed to convert protobuf models to gorm: %w", errConvertPB)
	}

	documentORM, err := md.Document.ToORM(ctx)
	if err != nil {
		return documentORM, fmt.Errorf("%w", err)
	}

	return documentORM, nil
}

func parseFile(fileName string) *protobom.Document {
	document, err := sbomReader.ParseFile(fileName)
	if err != nil {
		return nil
	}

	return document
}

func TestAddDocument(t *testing.T) {
	var err error
	db, err = CreateSchema(":memory:")
	if err != nil {
		t.FailNow()
	}

	ctx = context.Background()
	cdx := &mockDocument{Document: *parseFile("testdata/sbom.cdx.json")}
	cdxError := &mockDocument{Document: *parseFile("testdata/sbom.cdx.json"), shouldError: true}
	spdx := &mockDocument{Document: *parseFile("testdata/sbom.spdx.json")}

	for _, data := range []struct {
		document      PBToORMConverter
		expectedError string
		name          string
	}{
		{
			document: cdx,
			name:     "valid CycloneDX document",
		},
		{
			document: spdx,
			name:     "valid SPDX document",
		},
		{
			document:      cdxError,
			expectedError: "failed to convert protobuf models to gorm: %w",
			name:          "uninitialized struct fields",
		},
	} {
		t.Run(data.name, func(t *testing.T) {
			err := AddDocument(data.document.(*mockDocument))
			if data.expectedError != "" {
				require.Errorf(t, err, data.expectedError, err)
				return
			}

			require.Nil(t, err)
		})
	}
}

func TestCreateSchema(t *testing.T) {
	for _, data := range []struct {
		expectedDB    *gorm.DB
		expectedDSN   string
		expectedError string
		dbFile        string
		name          string
	}{
		{
			name:          "in-memory database",
			dbFile:        ":memory:",
			expectedDSN:   ":memory:?_pragma=foreign_keys(1)",
			expectedError: "",
		},
		{
			name:          "nonexistent path",
			dbFile:        "/missing/path/to/bomctl.db",
			expectedDSN:   "",
			expectedError: "error opening database file %s: %w",
		},
	} {
		t.Run(data.name, func(t *testing.T) {
			db = nil
			db, err := CreateSchema(data.dbFile)

			if data.expectedDSN != "" {
				require.Equal(t, data.expectedDSN, db.Dialector.(*sqlite.Dialector).DSN)

				// Test idempotence.
				db2, err := CreateSchema(data.dbFile)
				if err != nil {
					t.FailNow()
				}

				require.Equal(t, db, db2)
			} else {
				require.Nil(t, db)
				require.Errorf(t, err, data.expectedError, data.dbFile, err)
			}
		})
	}
}
