// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/fetch.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/fetch"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const fetchMinArgs int = 1

func fetchCmd() *cobra.Command {
	opts := &options.FetchOptions{}
	outputFileName := outputFileValue("")

	fetchCmd := &cobra.Command{
		Use:   "fetch [flags] SBOM_URL...",
		Args:  cobra.MinimumNArgs(fetchMinArgs),
		Short: "Fetch SBOM file(s) from HTTP(S), OCI, Git, or Github URLs",
		Long:  "Fetch SBOM file(s) from HTTP(S), OCI, Git, or Github URLs",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			clientString, err := cmd.Flags().GetString("client")
			cobra.CheckErr(err)
			opts.Client = clientString

			if outputFileName != "" {
				if len(args) > 1 {
					opts.Logger.Fatal("The --output-file option cannot be used when more than one URL is provided.")
				}

				out, err := os.Create(string(outputFileName))
				if err != nil {
					opts.Logger.Fatal("error creating output file", "outputFileName", outputFileName)
				}

				opts.OutputFile = out

				defer opts.OutputFile.Close()
			}

			for _, url := range args {
				if _, err := fetch.Fetch(url, opts); err != nil {
					opts.Logger.Fatal(err)
				}
			}
		},
	}

	fetchCmd.Flags().VarP(&outputFileName, "output-file", "o", "Path to output file")
	fetchCmd.Flags().BoolVar(&opts.UseNetRC, "netrc", false, "Use .netrc file for authentication to remote hosts")
	fetchCmd.Flags().StringVar(&opts.Alias, "alias", "", "Readable identifier to apply to document")
	fetchCmd.Flags().StringVar(&opts.Client, "client", "", "Specify client type to use for fetch")
	fetchCmd.Flags().StringArrayVar(&opts.Tags, "tag", []string{},
		"Tag(s) to apply to document (can be specified multiple times)")

	return fetchCmd
}
