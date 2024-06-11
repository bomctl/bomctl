// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/save.go
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

	"github.com/bomctl/bomctl/internal/pkg/save"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func saveCmd() *cobra.Command {
	saveCmd := &cobra.Command{
		Use:    "save [flags] SBOM_ID...",
		Args:   cobra.MinimumNArgs(1),
		PreRun: parseSavePositionalArgs,
		Short:  "Save SBOM file(s) from Storage",
		Long:   "Save SBOM file(s) from Storage to Filesystem",
		Run: func(_ *cobra.Command, _ []string) {
			var err error
			logger = utils.NewLogger("save")

			for _, sbomID := range sbomIDs {
				if err = save.Exec(sbomID, outputFile.String(), format, encoding); err != nil {
					logger.Error(err)
				}
			}

			if err != nil {
				os.Exit(1)
			}
		},
	}

	saveCmd.Flags().StringVarP(
		&format,
		"format",
		"f",
		"",
		"output format [spdx, spdx-2.3, cyclonedx, cyclonedx-1.0, cyclonedx-1.1, cyclonedx-1.2, cyclonedx-1.3, cyclonedx-1.4, cyclonedx-1.5]")

	saveCmd.Flags().StringVarP(
		&encoding,
		"encoding",
		"e",
		"json",
		"the output encoding [spdx: [text, json] cyclonedx: [json]")

	return saveCmd
}

func parseSavePositionalArgs(_ *cobra.Command, args []string) {
	for _, arg := range args {
		sbomIDs = append(sbomIDs, arg)
	}
}
