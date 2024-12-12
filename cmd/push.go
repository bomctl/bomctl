// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/push.go
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
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/push"
)

func pushCmd() *cobra.Command {
	opts := &options.PushOptions{}

	pushCmd := &cobra.Command{
		Use:   "push [flags] SBOM_ID DEST_PATH",
		Args:  cobra.MinimumNArgs(2),
		Short: "Push stored SBOM file to remote URL or filesystem",
		Long:  "Push stored SBOM file to remote URL or filesystem",
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			formatString := cmd.Flag("format").Value.String()
			encoding := cmd.Flag("encoding").Value.String()

			format, err := parseFormat(formatString, encoding)
			if err != nil {
				opts.Logger.Fatal(err, "format", formatString, "encoding", encoding)
			}

			opts.Format = format

			// Get the document to obtain its ID, in case the provided ID was an alias.
			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				opts.Logger.Fatal(err, "documentID", args[0])
			}

			if err := push.Push(document.GetMetadata().GetId(), args[1], opts); err != nil {
				opts.Logger.Fatal(err)
			}
		},
	}

	formatValue, encodingValue := formatChoice(), encodingChoice()

	pushCmd.Flags().VarP(formatValue, "format", "f", formatValue.Usage())
	pushCmd.Flags().VarP(encodingValue, "encoding", "e", encodingValue.Usage())
	pushCmd.Flags().BoolVar(&opts.UseNetRC, "netrc", false, "Use .netrc file for authentication to remote hosts")
	pushCmd.Flags().BoolVar(&opts.UseTree, "tree", false, "Recursively push all SBOMs in external reference tree")

	return pushCmd
}
