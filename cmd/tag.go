// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/tag.go
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
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	tagAddMinArgs     = 2
	tagClearExactArgs = 1
	tagListExactArgs  = 1
	tagRemoveMinArgs  = 2
)

func tagCmd() *cobra.Command {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Edit the tags of a document",
		Long:  "Edit the tags of a document",
	}

	tagCmd.AddCommand(tagAddCmd(), tagClearCmd(), tagListCmd(), tagRemoveCmd())

	return tagCmd
}

func tagAddCmd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add [flags] SBOM_ID TAGS...",
		Short: "Add tags to a document",
		Long:  "Add tags to a document",
		Args:  cobra.MinimumNArgs(tagAddMinArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			if err := backend.AddAnnotations(document.GetMetadata().GetId(),
				db.TagAnnotation, args[1:]...); err != nil {
				backend.Logger.Fatalf("failed to add tags: %v", err)
			}
		},
	}

	return addCmd
}

func tagClearCmd() *cobra.Command {
	clearCmd := &cobra.Command{
		Use:   "clear [flags] SBOM_ID...",
		Short: "Clear all tags from a document",
		Long:  "Clear all tags from a document",
		Args:  cobra.ExactArgs(tagClearExactArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			tagsToRemove, err := backend.GetDocumentTags(document.GetMetadata().GetId())
			if err != nil {
				backend.Logger.Fatalf("failed to clear tags: %v", err)
			}

			err = backend.RemoveAnnotations(document.GetMetadata().GetId(), db.TagAnnotation, tagsToRemove...)
			if err != nil {
				backend.Logger.Fatalf("failed to clear tags: %v", err)
			}
		},
	}

	return clearCmd
}

func tagListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags] SBOM_ID",
		Aliases: []string{"ls"},
		Short:   "List the tags of a document",
		Long:    "List the tags of a document",
		Args:    cobra.ExactArgs(tagListExactArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatal("Failed to get document", "err", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			tags, err := backend.GetDocumentTags(document.GetMetadata().GetId())
			if err != nil {
				backend.Logger.Fatal("Failed to get document tags", "err", err)
			}

			sort.Strings(tags)

			fmt.Fprintf(os.Stdout, "\nTags for %v\n%v\n", args[0], strings.Repeat("─", cliTableWidth))
			fmt.Fprintf(os.Stdout, "%v\n\n", strings.Join(tags, "\n"))
		},
	}

	return listCmd
}

func tagRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] SBOM_ID TAGS...",
		Aliases: []string{"rm"},
		Short:   "Remove specified tags from a document",
		Long:    "Remove specified tags from a document",
		Args:    cobra.MinimumNArgs(tagRemoveMinArgs),
		Run: func(cmd *cobra.Command, args []string) {
			backend := backendFromContext(cmd)

			defer backend.CloseClient()

			document, err := backend.GetDocumentByIDOrAlias(args[0])
			if err != nil {
				backend.Logger.Fatalf("failed to get document: %v", err)
			}

			if document == nil {
				backend.Logger.Fatal(errDocumentNotFound)
			}

			err = backend.RemoveAnnotations(document.GetMetadata().GetId(), db.TagAnnotation, args[1:]...)
			if err != nil {
				backend.Logger.Fatalf("failed to remove tags: %v", err)
			}
		},
	}

	return removeCmd
}
