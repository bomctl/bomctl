// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/main_test.go
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

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/bomctl/bomctl/cmd"
	"github.com/bomctl/bomctl/internal/e2e/e2eutil"
	"github.com/bomctl/bomctl/internal/pkg/db"
)

func TestMain(m *testing.M) {
	exitCode := testscript.RunMain(m, map[string]func() int{
		"bomctl": cmd.Execute,
	})
	os.Exit(exitCode)
}

func TestBomctlList(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:                 "testData/list",
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			tmpDir, err := os.MkdirTemp("", "list-cmd-test")
			if err != nil {
				t.Fatalf("%v", err)
			}

			env.Setenv("TMP_DIR", tmpDir)

			backend, err := db.NewBackend(db.WithDatabaseFile(filepath.Join(tmpDir, db.DatabaseFile)))
			if err != nil {
				t.Fatalf("%v", err)
			}

			testdataDir := filepath.Join("..", "pkg", "db", "testdata")

			sboms, err := os.ReadDir(testdataDir)
			if err != nil {
				t.Fatalf("%v", err)
			}

			for sbomIdx := range sboms {
				sbomData, err := os.ReadFile(filepath.Join(testdataDir, sboms[sbomIdx].Name()))
				if err != nil {
					t.Fatalf("%v", err)
				}

				_, err = backend.AddDocument(sbomData)
				if err != nil {
					t.Fatalf("failed storing document: %v", err)
				}
			}

			return nil
		},
	})
}

func TestBomctlFetch(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:                 "testData/fetch",
		Condition:           e2eutil.CustomConditions,
		Cmds:                e2eutil.CustomCommands(),
		RequireExplicitExec: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv("HTTPS_PROXY", os.Getenv("HTTPS_PROXY"))

			return nil
		},
	})
}
