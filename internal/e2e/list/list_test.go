// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/list/list_test.go
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

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/e2e"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type ListCmdTestSuite struct {
	suite.Suite
	*options.Options
	*db.Backend
	*e2e.E2E
	tmpDir string
}

func (lcts *ListCmdTestSuite) SetupSuite() {
	var err error

	lcts.tmpDir, err = os.MkdirTemp("", "list-cmd-test")
	lcts.Require().NoErrorf(err, "Failed to create temporary directory: %v", err)

	lcts.Backend, err = db.NewBackend(db.WithDatabaseFile(filepath.Join(lcts.tmpDir, db.DatabaseFile)))
	lcts.Require().NoError(err)

	testdataDir := filepath.Join("..", "..", "pkg", "db", "testdata")

	sboms, err := os.ReadDir(testdataDir)
	if err != nil {
		lcts.T().Fatalf("%v", err)
	}

	for sbomIdx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join(testdataDir, sboms[sbomIdx].Name()))
		if err != nil {
			lcts.T().Fatalf("%v", err)
		}

		_, err = lcts.Backend.AddDocument(sbomData)
		if err != nil {
			lcts.FailNow("failed storing document", "err", err)
		}
	}

	lcts.E2E = e2e.NewE2E(lcts.T())
}

func (lcts *ListCmdTestSuite) TearDownSuite() {
	if err := os.RemoveAll(lcts.tmpDir); err != nil {
		lcts.T().Fatalf("Error removing temp directory %s", db.DatabaseFile)
	}
}

func (lcts *ListCmdTestSuite) TestListNoInput() {
	testFilePath := filepath.Join("..", "testData", "list_no_input.txt")

	expectedOutput, err := os.ReadFile(testFilePath)
	if err != nil {
		lcts.Fatalf("Failed to read test file: %f", err)
	}

	stdOut, stdErr, err := lcts.E2E.Bomctl(lcts.T(), "list", "--cache-dir", lcts.tmpDir)

	lcts.Require().NoError(err, stdOut, stdErr)
	lcts.Require().Empty(stdErr)
	lcts.Require().Equal(string(expectedOutput), stdOut)
}

func TestListCommandSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ListCmdTestSuite))
}
