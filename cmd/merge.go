// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/merge.go
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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/merge"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func mergeCmd() *cobra.Command {
	opts := &options.MergeOptions{}

	mergeCmd := &cobra.Command{
		Use:   "merge [flags] DOCUMENT_ID...",
		Args:  cobra.MinimumNArgs(1),
		Short: "Merge SBOM documents in local storage",
		Long: fmt.Sprintf("%s%s%s",
			"Merge SBOM documents in local storage. The leftmost specified document ID takes priority with the ",
			"intent of only updating the field if the existing value is empty. Lists are de-duplicated based off of ",
			"a combination of fields depending on the type",
		),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)
			backend.Logger.SetPrefix("merge")

			defer backend.CloseClient()

			documentName, err := cmd.Flags().GetString("name")
			cobra.CheckErr(err)

			opts.DocumentName = documentName

			if _, err := merge.Merge(args, opts); err != nil {
				backend.Logger.Fatal(err)
			}
		},
	}

	mergeCmd.Flags().StringP("name", "n", "", "Name of merged document")
	mergeCmd.Flags().StringVar(&opts.Alias, "alias", "", "Readable identifier to apply to merged document")
	mergeCmd.Flags().StringArrayVar(&opts.Tags, "tag", []string{},
		"Tag(s) to apply to merged document (can be specified multiple times)")

	return mergeCmd
}
