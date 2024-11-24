// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/db/db_revised_test.go
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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/testutil"
)

type dbrSuite struct {
	suite.Suite
	*db.Backend
	documentInfo []testutil.DocumentInfo
}

func (dbrs *dbrSuite) SetupSubTest() {
	var err error

	dbrs.Backend, err = testutil.NewTestBackend()
	dbrs.Require().NoError(err, "failed database backend creation")

	dbrs.documentInfo, err = testutil.AddTestDocuments(dbrs.Backend)
	dbrs.Require().NoError(err, "failed database backend setup")
}

func (dbrs *dbrSuite) TearDownSubTest() {
	dbrs.Backend.CloseClient()
}

func (dbrs *dbrSuite) TestBackend_AddDocumentRevision() {
	for _, data := range []struct {
		name     string
		baseID   string
		alias    string
		errorMsg string
	}{
		{
			name:     "existing alias on base doc",
			baseID:   "8daeb29e-8655-fae1-b792-78b998823fc6",
			alias:    "cdx",
			errorMsg: "",
		},
	} {
		dbrs.Run(data.name, func() {
			err := dbrs.Backend.ClearDocumentAnnotations(dbrs.documentInfo[1].Document.GetMetadata().GetId())
			dbrs.Require().NoError(err)

			docContent := dbrs.documentInfo[1].Content

			newDoc, err := dbrs.Backend.AddRevisedDocument(dbrs.documentInfo[0].Document, docContent)

			if data.errorMsg == "" {
				dbrs.Require().NoError(err)

				newID := newDoc.GetMetadata().GetId()
				baseDocID := dbrs.documentInfo[0].Document.GetMetadata().GetId()

				baseID, err := dbrs.Backend.GetDocumentUniqueAnnotation(newID, db.BaseDocumentAnnotation)
				dbrs.Require().NoError(err)
				dbrs.Require().Equal(data.baseID, baseID)

				alias, err := dbrs.Backend.GetDocumentUniqueAnnotation(newID, db.AliasAnnotation)
				dbrs.Require().NoError(err)
				dbrs.Require().Equal(data.alias, alias)

				alias, err = dbrs.Backend.GetDocumentUniqueAnnotation(baseDocID, db.AliasAnnotation)
				dbrs.Require().NoError(err)
				dbrs.Require().Equal("", alias)

				soureData, err := dbrs.Backend.GetDocumentUniqueAnnotation(newID, db.SourceDataAnnotation)
				dbrs.Require().NoError(err)
				dbrs.Require().Equal("", soureData)
			} else {
				dbrs.Require().EqualError(err, data.errorMsg)
			}
		})
	}
}

func (dbrs *dbrSuite) TestBackend_UpdateAliasReference() {
	baseID := dbrs.documentInfo[1].Document.GetMetadata().GetId()
	revisedID := dbrs.documentInfo[0].Document.GetMetadata().GetId()

	for _, data := range []struct {
		prep     func()
		cleanup  func()
		name     string
		alias    string
		errorMsg string
	}{
		{
			name: "No existing alias on base doc",
			prep: func() {
				err := dbrs.Backend.RemoveDocumentAnnotations(revisedID, db.AliasAnnotation)
				dbrs.Require().NoError(err)

				err = dbrs.Backend.RemoveDocumentAnnotations(baseID, db.AliasAnnotation)
				dbrs.Require().NoError(err)
			},
			alias:    "",
			errorMsg: "",
		},
		{
			name: "existing alias on base doc",
			prep: func() {
				err := dbrs.Backend.RemoveDocumentAnnotations(revisedID, db.AliasAnnotation)
				dbrs.Require().NoError(err)
			},
			alias:    "spdx",
			errorMsg: "",
		},
		{
			name: "existing alias on revised doc",
			prep: func() {
			},
			errorMsg: "failed to set alias: the document already has an alias",
		},
	} {
		dbrs.Run(data.name, func() {
			data.prep()

			err := dbrs.Backend.UpdateAliasReference(dbrs.documentInfo[1].Document, dbrs.documentInfo[0].Document)
			if data.errorMsg == "" {
				dbrs.Require().NoError(err)
				docAlias, err := dbrs.Backend.GetDocumentUniqueAnnotation(revisedID, db.AliasAnnotation)
				dbrs.Require().NoError(err)
				dbrs.Require().Equal(data.alias, docAlias)
			} else {
				dbrs.Require().EqualError(err, data.errorMsg)
			}
		})
	}
}

func TestDBRSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(dbrSuite))
}
