// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/fetch.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
package cmd

import "github.com/spf13/cobra"

func fetchCmd() *cobra.Command {
	fetchCmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:  "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Run:   fetch,
	}

	fetchCmd.Flags().VarP(
		&sbomUrls,
		"sbom-url",
		"u",
		"URL of SBOM to fetch (can be specified multiple times)",
	)
	fetchCmd.Flags().VarP(
		&outputFile,
		"output-file",
		"o",
		"Path to output file [default: hopctl-merge-YYMMDD-HHMMSS.json]",
	)

	return fetchCmd
}

func fetch(cmd *cobra.Command, args []string) {}
