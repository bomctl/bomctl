// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/setup_cache.go
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

package e2eutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	setupCacheMinArgNum = 1
	setupCacheMaxArgNum = 2
)

// setupCache takes the work dir from the test environment
// and an optional string to match against existing test files
// and populates the cache with expected test files to be used in the test script.
//
// ex: setup_cache <WORK> <fileMatch>
//
//	^ This will create a DB file,and populate it with files that match the included string
//
// ex: setup_cache <WORK>
//
//	^ This will create a DB file, and populate it will all files in testdata dir
func setupCache(script *testscript.TestScript, _ bool, args []string) {
	if len(args) < setupCacheMinArgNum || len(args) > setupCacheMaxArgNum {
		script.Fatalf("syntax: setup_cache work_dir file_match")
	}

	workDir := args[0]
	fileMatch := ""

	if len(args) == setupCacheMaxArgNum {
		fileMatch = args[1]
	}

	backend, err := db.NewBackend(db.WithDatabaseFile(filepath.Join(workDir, db.DatabaseFile)))
	script.Check(err)

	testDataDir := filepath.Join("..", "testdata")

	sboms, err := os.ReadDir(testDataDir)
	script.Check(err)

	for sbomIdx := range sboms {
		name := sboms[sbomIdx].Name()
		if !strings.Contains(name, fileMatch) {
			continue
		}

		sbomData, err := os.ReadFile(filepath.Join(testDataDir, name))
		script.Check(err)

		_, err = backend.AddDocument(sbomData, db.WithSourceDocumentAnnotations(sbomData))
		script.Check(err)
	}
}
