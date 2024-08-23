// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
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

	"github.com/bomctl/bomctl/internal/pkg/logger"
)

//nolint:gochecknoglobals
var (
	// BuildDate is the date and time this binary was built.
	BuildDate string

	// Version is the full version string.
	Version = fmt.Sprintf("v%s.%s.%s%s", VersionMajor, VersionMinor, VersionPatch, VersionPre)

	// VersionMajor is for breaking API changes.
	VersionMajor string

	// VersionMinor is for functionality in a backwards-compatible manner.
	VersionMinor string

	// VersionPatch is for backwards-compatible bug fixes.
	VersionPatch string

	// VersionPre indicates prerelease branch.
	VersionPre string

	// VersionString is the specification version that the package types support.
	VersionString = fmt.Sprintf("%s (built on %s)", Version, BuildDate)
)

func versionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  "Print the version",
		Run: func(_ *cobra.Command, _ []string) {
			logger.New("bomctl").Print("", "version", Version, "buildDate", BuildDate)
		},
	}

	return versionCmd
}
