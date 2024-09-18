// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/list.go
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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/format"
)

func listCmd() *cobra.Command {
	tags := []string{}

	listCmd := &cobra.Command{
		Use:     "list [flags] SBOM_ID...",
		Aliases: []string{"ls"},
		Short:   "List SBOM documents in local cache",
		Long:    "List SBOM documents in local cache",
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)
			backend.Logger.SetPrefix("list")

			defer backend.CloseClient()

			documents, err := backend.GetDocumentsByIDOrAlias(args...)
			if err != nil {
				backend.Logger.Fatalf("failed to get documents: %v", err)
			}

			if len(tags) > 0 {
				documents, err = backend.FilterDocumentsByTag(documents, tags...)
				if err != nil {
					backend.Logger.Fatalf("failed to get documents: %v", err)
				}
			}

			listOutput := format.NewTable()
			for _, document := range documents {
				listOutput.AddRow(document, backend)
			}

			fmt.Fprintln(os.Stdout, listOutput.String())
		},
	}

	listCmd.Flags().StringArrayVar(&tags, "tag", []string{},
		"Tag(s) used to filter documents (can be specified multiple times)")

	return listCmd
}
