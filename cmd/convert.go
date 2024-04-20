// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/convert.go
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

	"github.com/bomctl/bomctl/internal/pkg/convert"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func convertCmd() *cobra.Command {
	convertCmd := &cobra.Command{
		Use:    "convert [flags] SBOM_URL...",
		Args:   cobra.MinimumNArgs(1),
		PreRun: parsePositionalArgs,
		Short:  "Convert SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:   "Convert SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Run: func(_ *cobra.Command, _ []string) {
			var err error
			logger = utils.NewLogger("convert")

			for _, url := range sbomURLs {
				err = convert.Exec(url,
					outputFile.String(),
					outputBomFormat.String(),
					outputBomEncoding.String(),
					useNetRC)

				if err != nil {
					logger.Error(err)
				}
			}

			if err != nil {
				os.Exit(1)
			}
		},
	}

	convertCmd.Flags().VarP(
		&outputFile,
		"output-file",
		"o",
		"Path to output file",
	)
	convertCmd.Flags().BoolVar(
		&useNetRC,
		"netrc",
		false,
		"Use .netrc file for authentication to remote hosts",
	)
	convertCmd.Flags().VarP(
		&outputBomFormat,
		"format",
		"f",
		"the output format [spdx, spdx-2.3, cyclonedx, cyclonedx-1.0, cyclonedx-1.1, cyclonedx-1.2, cyclonedx-1.3, cyclonedx-1.4, cyclonedx-1.5]",
	)
	convertCmd.Flags().VarP(
		&outputBomEncoding,
		"encoding",
		"e",
		"the output encoding [spdx: [text, json] cyclonedx: [json]",
	)

	// cmd.Flags().StringVarP(&o.Format, "format", "f", "", "the output format [spdx, spdx-2.3, cyclonedx, cyclonedx-1.0, cyclonedx-1.1, cyclonedx-1.2, cyclonedx-1.3, cyclonedx-1.4, cyclonedx-1.5]") //nolint: lll
	// cmd.Flags().StringVarP(&o.Encoding, "encoding", "e", "json", "the output encoding [spdx: [text, json] cyclonedx: [json]")

	return convertCmd
}
