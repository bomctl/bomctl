// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/check_files.go
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
	"path"

	"github.com/rogpeppe/go-internal/testscript"
)

// checkFile is a testscript command that checks the
// existence of a list of files inside a directory.
func checkFiles(script *testscript.TestScript, neg bool, args []string) { //nolint: revive
	if len(args) < 1 {
		script.Fatalf("syntax: check_file directory_name file_name [file_name ...]")
	}

	dir := args[0]

	for i := 1; i < len(args); i++ {
		file := path.Join(dir, args[i])
		if neg {
			if fileExists(file) {
				script.Fatalf("file %s found", file)
			}
		}

		if !fileExists(file) {
			script.Fatalf("file not found %s", file)
		}
	}
}
