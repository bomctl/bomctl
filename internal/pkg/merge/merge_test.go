// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/merge/merge_test.go
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

package merge_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/merge"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

type mergeSuite struct {
	suite.Suite
	*options.Options
	*db.Backend
	documentInfo []testutil.DocumentInfo
}

func (ms *mergeSuite) SetupSuite() {
	var err error

	ms.Backend, err = testutil.NewTestBackend()
	ms.Require().NoError(err, "failed database backend creation")

	ms.documentInfo, err = testutil.AddTestDocuments(ms.Backend)
	ms.Require().NoError(err, "failed database backend setup")

	ms.Options = options.New().WithContext(context.WithValue(context.Background(), db.BackendKey{}, ms.Backend))
}

func (ms *mergeSuite) TearDownSuite() {
	ms.Backend.CloseClient()
}

func (ms *mergeSuite) TestMerge() {
	opts := &options.MergeOptions{
		Options: ms.Options,
	}

	baseDocument := ms.documentInfo[0].Document
	otherDocument := ms.documentInfo[1].Document

	docID, err := merge.Merge([]string{baseDocument.GetMetadata().GetId(), otherDocument.GetMetadata().GetId()}, opts)
	ms.Require().NoError(err)

	mergedDocument, err := ms.Backend.GetDocumentByID(docID)
	ms.Require().NoError(err, "Failed to get merged document from DB")

	if baseDocument.GetMetadata().GetName() != "" {
		ms.Equal(baseDocument.GetMetadata().GetName(), mergedDocument.GetMetadata().GetName())
	} else if otherDocument.GetMetadata().GetName() != "" {
		ms.Equal(otherDocument.GetMetadata().GetName(), mergedDocument.GetMetadata().GetName())
	}

	mergedTools := append(baseDocument.GetMetadata().GetTools(), otherDocument.GetMetadata().GetTools()...)

	ms.Len(mergedDocument.GetMetadata().GetTools(), len(mergedTools))
}

func TestMergeSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(mergeSuite))
}
