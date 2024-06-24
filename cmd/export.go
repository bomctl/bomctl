// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: cmd/export.go
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
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/export"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func exportCmd() *cobra.Command {
	opts := &export.ExportOptions{
		Logger: utils.NewLogger("export"),
	}

	outputFile := OutputFileValue("")
	formatString := FormatStringValue("")
	formatEncoding := FormatEncodingValue("json")

	sbomIDs := SBOMIDSliceValue{}

	exportCmd := &cobra.Command{
		Use:  "export [flags] SBOM_URL...",
		Args: cobra.MinimumNArgs(1),
		PreRun: func(_ *cobra.Command, args []string) {
			sbomIDs = append(sbomIDs, args...)
		},
		Short: "Export SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:  "Export SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Run: func(_ *cobra.Command, _ []string) {
			opts.CacheDir = viper.GetString("cache_dir")
			opts.ConfigFile = viper.GetString("config_file")
			opts.FormatString = string(formatString)
			opts.Encoding = string(formatEncoding)

			if string(outputFile) != "" {
				if len(sbomIDs) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one SBOM  is provided.")
				}

				out, err := os.Create(string(outputFile))
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}

			for _, id := range sbomIDs {
				if err := export.Export(id, opts); err != nil {
					opts.Logger.Fatal(err)
				}
			}
		},
	}

	exportCmd.Flags().VarP(
		&outputFile,
		"output-file",
		"o",
		"Path to output file",
	)
	exportCmd.Flags().VarP(
		&formatString,
		"format",
		"f",
		"output format [spdx, spdx-2.3, cyclonedx, cyclonedx-1.0, cyclonedx-1.1, cyclonedx-1.2, cyclonedx-1.3, cyclonedx-1.4, cyclonedx-1.5]")

	exportCmd.Flags().VarP(
		&formatEncoding,
		"encoding",
		"e",
		"the output encoding [spdx: [text, json] cyclonedx: [json]")

	return exportCmd
}

// func exportCmd() *cobra.Command {
// 	exportCmd := &cobra.Command{
// 		Use:    "export [flags] SBOM_ID...",
// 		Args:   cobra.MinimumNArgs(1),
// 		PreRun: parseExportPositionalArgs,
// 		Short:  "Export SBOM file(s) from Storage",
// 		Long:   "Export SBOM file(s) from Storage to Filesystem",
// 		Run: func(_ *cobra.Command, _ []string) {
// 			var err error
// 			logger = utils.NewLogger("export")

// 			for _, sbomID := range sbomIDs {
// 				if err = export.Exec(sbomID, outputFile.String(), format, encoding); err != nil {
// 					logger.Error(err)
// 				}
// 			}

// 			if err != nil {
// 				os.Exit(1)
// 			}
// 		},
// 	}

// 	exportCmd.Flags().StringVarP(
// 		&format,
// 		"format",
// 		"f",
// 		"",
// 		"output format [spdx, spdx-2.3, cyclonedx, cyclonedx-1.0, cyclonedx-1.1, cyclonedx-1.2, cyclonedx-1.3, cyclonedx-1.4, cyclonedx-1.5]")

// 	exportCmd.Flags().StringVarP(
// 		&encoding,
// 		"encoding",
// 		"e",
// 		"json",
// 		"the output encoding [spdx: [text, json] cyclonedx: [json]")

// 	return exportCmd
// }

// func parseExportPositionalArgs(_ *cobra.Command, args []string) {
// 	for _, arg := range args {
// 		sbomIDs = append(sbomIDs, arg)
// 	}
// }
