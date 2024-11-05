// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/add_files.go
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

	"github.com/bomctl/bomctl/internal/testutil"
)

const (
	addFilesMinArgNum = 1
	addFilesMaxArgNum = 2
	addFilesFilePerm  = 0o600
)

// addFiles takes the work dir from the test environment
// and an optional string to match against existing test files
// and moves the test files into the test env for the script to use.

// ex: add_files <WORK>
//
//	^ This will write all available files to the test env working dir
//
// ex: add_files <WORK> <fileMatch>
//
//	^ This will write all files that match <fileMatch> to test working dir

func addFiles(script *testscript.TestScript, _ bool, args []string) {
	if len(args) < addFilesMinArgNum || len(args) > addFilesMaxArgNum {
		script.Fatalf("syntax: add_files work_dir file_match")
	}

	workDir := args[0]
	fileMatch := ""

	if len(args) == addFilesMaxArgNum {
		fileMatch = args[1]
	}

	testDataDir := testutil.GetTestdataDir()

	sboms, err := os.ReadDir(testDataDir)
	script.Check(err)

	for sbomIdx := range sboms {
		name := sboms[sbomIdx].Name()
		if !strings.Contains(name, fileMatch) {
			continue
		}

		sbomData, err := os.ReadFile(filepath.Join(testDataDir, name))
		script.Check(err)

		err = os.WriteFile(filepath.Join(workDir, name), sbomData, addFilesFilePerm)
		script.Check(err)
	}
}
