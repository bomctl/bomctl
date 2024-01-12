// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/fetch.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/pkg/utils"
)

func fetchCmd() *cobra.Command {
	fetchCmd := &cobra.Command{
		Use:    "fetch [flags] SBOM_URL...",
		Args:   cobra.MinimumNArgs(1),
		PreRun: parsePositionalArgs,
		Short:  "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:   "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Run:    fetch,
	}

	fetchCmd.Flags().VarP(
		&outputFile,
		"output-file",
		"o",
		"Path to output file",
	)

	return fetchCmd
}

func fetch(cmd *cobra.Command, args []string) {
	for _, url := range sbomUrls {
		parsedURL := utils.ParseURL(url)

		switch parsedURL.Scheme {
		case "git":
		case "http", "https":
			cobra.CheckErr(utils.DownloadHTTP(url, outputFile.String(), nil))
		case "oci":
		}
	}
}

func parsePositionalArgs(cmd *cobra.Command, args []string) {
	for _, arg := range args {
		sbomUrls = append(sbomUrls, arg)
	}
}
