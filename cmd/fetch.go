// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
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
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

func fetchCmd() *cobra.Command {
	opts := &options.FetchOptions{
		Options: options.New(options.WithLogger(utils.NewLogger("fetch"))),
	}

	outputFile := outputFileValue("")

	fetchCmd := &cobra.Command{
		Use:    "fetch [flags] SBOM_URL...",
		Args:   cobra.MinimumNArgs(1),
		Short:  "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		Long:   "Fetch SBOM file(s) from HTTP(S), OCI, or Git URLs",
		PreRun: preRun(opts.Options),
		Run: func(_ *cobra.Command, args []string) {
			if string(outputFile) != "" {
				if len(args) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one URL is provided.")
				}

				out, err := os.Create(string(outputFile))
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFile", outputFile)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}

			for _, url := range args {
				if err := fetch.Fetch(url, opts); err != nil {
					opts.Logger.Fatal(err)
				}
			}
		},
	}

	fetchCmd.Flags().VarP(&outputFile, "output-file", "o", "Path to output file.")
	fetchCmd.Flags().BoolVarP(&opts.UseNetRC, "netrc", "n", false, "Use .netrc file for authentication to remote hosts.")
	fetchCmd.Flags().StringVarP(&opts.Alias, "alias", "a", "", "Alias used to identify the fetched document in a readable manner.")
	fetchCmd.Flags().StringArrayVarP(&opts.Tags, "tag", "t", []string{}, "Tag used to mark the fetched document. Can be specified multiple times to add multiple tags.")

	return fetchCmd
}
