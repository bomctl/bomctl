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
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/fetch"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func fetchCmd() *cobra.Command {
	opts := &fetch.FetchOptions{
		Logger:   utils.NewLogger("fetch"),
		UseNetRC: false,
	}

	outputFile := OutputFileValue("")
	sbomURLs := URLSliceValue{}

	fetchCmd := &cobra.Command{
		Use:   "fetch [flags] SBOM_URL...",
		Args:  cobra.MinimumNArgs(1),
		Short: "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:  "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		PreRun: func(_ *cobra.Command, args []string) {
			for _, arg := range args {
				sbomURLs = append(sbomURLs, arg)
			}
		},
		Run: func(_ *cobra.Command, _ []string) {
			opts.CacheDir = viper.GetString("cache_dir")
			opts.ConfigFile = viper.GetString("config_file")

			if string(outputFile) != "" {
				if len(sbomURLs) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one URL is provided.")
				}

				out, err := os.Create(string(outputFile))
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}

			for _, url := range sbomURLs {
				if err := fetch.Fetch(url, opts); err != nil {
					opts.Logger.Fatal(err)
				}
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
		&opts.UseNetRC,
		"netrc",
		false,
		"Use .netrc file for authentication to remote hosts",
	)

	return fetchCmd
}
