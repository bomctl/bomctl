// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/e2eutil_commands.go
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

import "github.com/rogpeppe/go-internal/testscript"

type commandFunc = func(*testscript.TestScript, bool, []string)

func CustomCommands() map[string]commandFunc {
	return map[string]commandFunc{
		// compare_docs will check that two documents are equal
		// invoke as "compare_docs workdir file1 file2"
		// The command can be negated,
		// i.e. it will succeed if the given files are not equal
		// "! compare_docs workdir file1 file2"
		"compare_docs": compareDocuments,
		// setupCache takes the work dir from the test environment
		// and an optional string to match against existing test files
		// and populates the cache with expected test files to be used in the test script.
		"setup_cache": setupCache,
		// addFiles takes the work dir from the test environment
		// and a string to match against existing test files
		// and moves the test files into the test env for the script to use.
		"add_files": addFiles,
	}
}
