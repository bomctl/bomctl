// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/version.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// ------------------------------------------------------------------------
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
// ------------------------------------------------------------------------
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  "Print the version",
		Run:   version,
	}

	return versionCmd
}

const (
	// VersionMajor is for an API incompatible changes.
	VersionMajor = 0

	// VersionMinor is for functionality in a backwards-compatible manner.
	VersionMinor = 0

	// VersionPatch is for backwards-compatible bug fixes.
	VersionPatch = 1

	// VersionDev indicates development branch. Releases will be empty string.
	VersionDev = "-dev.1"
)

func getVersion() string {
	return fmt.Sprintf("%d.%d.%d%s", VersionMajor, VersionMinor, VersionPatch, VersionDev)
}

func version(cmd *cobra.Command, args []string) {
	fmt.Println("bomctl version", getVersion())
}
