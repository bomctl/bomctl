// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/alias.go
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
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const (
	aliasRemoveMinArgNum = 1
	aliasRemoveMaxArgNum = 2
	aliasSetExactArgNum  = 2
)

func aliasCmd() *cobra.Command {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Edit the alias for a document",
		Long:  "Edit the alias for a document",
	}

	aliasCmd.AddCommand(aliasListCmd(), aliasRemoveCmd(), aliasSetCmd())

	return aliasCmd
}

func aliasListCmd() *cobra.Command {
	aliasListCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"ls"},
		Short:   "List all alias definitions",
		Long:    "List all alias definitions",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			documents, err := backend.GetDocumentsByID()
			if err != nil {
				backend.Logger.Fatal(err)
			}

			aliasDefinitions := []string{}

			for _, doc := range documents {
				alias, err := backend.GetDocumentUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation)
				if err != nil {
					backend.Logger.Fatalf("failed to get alias: %v", err)
				}

				if alias != "" {
					aliasDefinitions = append(aliasDefinitions,
						fmt.Sprintf("%v → %v", alias, doc.GetMetadata().GetId()))
				}
			}

			sort.Strings(aliasDefinitions)

			fmt.Fprintf(os.Stdout, "\nAlias Definitions\n%v\n", strings.Repeat("─", cliTableWidth))
			fmt.Fprintf(os.Stdout, "%v\n\n", strings.Join(aliasDefinitions, "\n"))
		},
	}

	return aliasListCmd
}

func aliasRemoveCmd() *cobra.Command {
	aliasRemoveCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID",
		Aliases: []string{"rm"},
		Short:   "Remove the alias for a specific document",
		Long:    "Remove the alias for a specific document",
		Args:    cobra.RangeArgs(aliasRemoveMinArgNum, aliasRemoveMaxArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			docAlias, err := backend.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.AliasAnnotation)
			if err != nil {
				backend.Logger.Fatal(err, "documentID", args[0])
			}

			if err := backend.RemoveAnnotations(document.GetMetadata().GetId(),
				db.AliasAnnotation, docAlias); err != nil {
				backend.Logger.Fatal(err, "name", db.AliasAnnotation, "value", docAlias)
			}
		},
	}

	return aliasRemoveCmd
}

func aliasSetCmd() *cobra.Command {
	opts := &options.AliasOptions{}

	aliasSetCmd := &cobra.Command{
		Use:   "set [flags] SBOM_ID NEW_ALIAS",
		Short: "Set the alias for a specific document",
		Long:  "Set the alias for a specific document",
		Args:  cobra.ExactArgs(aliasSetExactArgNum),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Options = optionsFromContext(cmd)
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "documentID", args[0], "err", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			if err := backend.SetAlias(document.GetMetadata().GetId(), args[1], opts.Force); err != nil {
				if errors.Is(err, db.ErrDocumentAliasExists) {
					backend.Logger.Fatal(
						"The document already has an alias. To replace it, re-run the command with the --force flag")
				} else {
					backend.Logger.Fatal(err, "documentID", document.GetMetadata().GetId(), "alias", args[1])
				}
			}
		},
	}

	aliasSetCmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force replacing an existing alias, if there is one")

	return aliasSetCmd
}
