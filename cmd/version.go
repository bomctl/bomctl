// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/version.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
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
	fmt.Print(getVersion())
}
