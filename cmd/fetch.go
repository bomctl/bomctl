// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/fetch.go
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
	"os"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/fetch"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func fetchCmd() *cobra.Command {
	fetchCmd := &cobra.Command{
		Use:    "fetch [flags] SBOM_URL...",
		Args:   cobra.MinimumNArgs(1),
		PreRun: parsePositionalArgs,
		Short:  "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:   "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Run: func(_ *cobra.Command, _ []string) {
			var (
				err    error
				output *os.File
			)

			logger = utils.NewLogger("fetch")

			if string(outputFile) != "" {
				if len(sbomURLs) > 1 {
					logger.Fatal("The --output-file option cannot be used when more than one URL is provided.")
				}

				if output, err = os.Create(string(outputFile)); err != nil {
					logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				defer output.Close()
			}

			for _, url := range sbomURLs {
				if err = fetch.Fetch(url, output, useNetRC); err != nil {
					logger.Error(err)
				}
			}

			if err != nil {
				os.Exit(1)
			}
		},
	}

	fetchCmd.Flags().VarP(
		&outputFile,
		"output-file",
		"o",
		"Path to output file",
	)
	fetchCmd.Flags().BoolVar(
		&useNetRC,
		"netrc",
		false,
		"Use .netrc file for authentication to remote hosts",
	)

	return fetchCmd
}

func parsePositionalArgs(_ *cobra.Command, args []string) {
	for _, arg := range args {
		sbomURLs = append(sbomURLs, arg)
	}
}
