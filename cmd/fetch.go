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
		cobra.CheckErr(utils.FetchSBOM(url, outputFile.String()))
	}
}

func parsePositionalArgs(cmd *cobra.Command, args []string) {
	for _, arg := range args {
		sbomUrls = append(sbomUrls, arg)
	}
}
