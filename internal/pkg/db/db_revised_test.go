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
		prep     func()
		name     string
		baseID   string
		alias    string
		errorMsg string
	}{
		{
			name:     "existing alias on base doc",
			prep:     func() {},
			baseID:   "8daeb29e-8655-fae1-b792-78b998823fc6",
			alias:    "cdx",
			errorMsg: "",
		},
		{
			name: "no alias on base doc",
			prep: func() {
				err := dbrs.Backend.RemoveDocumentAnnotations(
					dbrs.documentInfo[0].Document.GetMetadata().GetId(),
					db.AliasAnnotation)
				dbrs.Require().NoError(err)
			},
			baseID:   "8daeb29e-8655-fae1-b792-78b998823fc6",
			alias:    "",
			errorMsg: "",
		},
	} {
		dbrs.Run(data.name, func() {
			docContent := dbrs.documentInfo[1].Content
			baseDoc := dbrs.documentInfo[0].Document

			data.prep()

			err := dbrs.Backend.ClearDocumentAnnotations(dbrs.documentInfo[1].Document.GetMetadata().GetId())
			dbrs.Require().NoError(err)

			newDoc, err := dbrs.Backend.AddDocument(docContent, db.WithRevisedDocumentAnnotations(baseDoc))

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

func TestDBRSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(dbrSuite))
}
