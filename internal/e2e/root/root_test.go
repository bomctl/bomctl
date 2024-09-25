// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/root/root_test.go
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

package root_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/e2e"
	"github.com/bomctl/bomctl/internal/pkg/db"
)

type RootCmdTestSuite struct {
	suite.Suite
	*e2e.E2E
	tmpDir string
}

func (rcts *RootCmdTestSuite) SetupSuite() {
	var err error

	rcts.tmpDir, err = os.MkdirTemp("", "root-cmd-test")
	rcts.Require().NoErrorf(err, "Failed to create temporary directory: %v", err)

	rcts.E2E = e2e.NewE2E(rcts.T())
}

func (rcts *RootCmdTestSuite) TearDownSuite() {
	if err := os.RemoveAll(rcts.tmpDir); err != nil {
		rcts.T().Fatalf("Error removing temp directory %s", db.DatabaseFile)
	}
}

func (rcts *RootCmdTestSuite) TestRootHelp() {
	stdOut, stdErr, err := rcts.E2E.Bomctl(rcts.T(), "--help")

	rcts.Require().NoError(err, stdOut, stdErr)
	rcts.Require().Empty(stdErr)
}

func TestRootCommandSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RootCmdTestSuite))
}
