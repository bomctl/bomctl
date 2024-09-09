// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/list.go
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
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/spf13/cobra"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	columnIdxID = iota
	columnIdxAlias
	columnIdxVersion
	columnIdxNumNodes

	columnWidthID       = 47
	columnWidthAlias    = 12
	columnWidthVersion  = 9
	columnWidthNumNodes = 9

	paddingHorizontal = 1
	paddingVertical   = 0

	rowHeaderIdx = 0
	rowMaxHeight = 1
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

			rows := [][]string{}
			for _, document := range documents {
				rows = append(rows, getRow(document, backend))
			}

			fmt.Fprintf(os.Stdout, "\n%s\n\n", table.New().
				Headers("ID", "Alias", "Version", "# Nodes").
				Rows(rows...).
				BorderTop(false).
				BorderBottom(false).
				BorderLeft(false).
				BorderRight(false).
				BorderHeader(true).
				StyleFunc(styleFunc).
				String(),
			)
		},
	}

	listCmd.Flags().StringArrayVar(&tags, "tag", []string{},
		"Tag(s) used to filter documents (can be specified multiple times)")

	return listCmd
}

func styleFunc(row, col int) lipgloss.Style {
	width := 0
	align := lipgloss.Center

	switch col {
	case columnIdxID:
		width = columnWidthID

		if row != rowHeaderIdx {
			align = lipgloss.Left
		}
	case columnIdxAlias:
		width = columnWidthAlias

		if row != rowHeaderIdx {
			align = lipgloss.Left
		}
	case columnIdxVersion:
		width = columnWidthVersion
	case columnIdxNumNodes:
		width = columnWidthNumNodes
	}

	return lipgloss.NewStyle().
		Padding(paddingVertical, paddingHorizontal).
		Width(width).
		AlignHorizontal(align).
		MaxHeight(rowMaxHeight)
}

func getRow(doc *sbom.Document, backend *db.Backend) []string {
	id := doc.GetMetadata().GetName()
	if id == "" {
		id = doc.GetMetadata().GetId()
	}

	alias, err := backend.GetDocumentUniqueAnnotation(doc.Metadata.Id, db.BomctlAnnotationAlias)
	if err != nil {
		backend.Logger.Fatalf("failed to get alias: %v", err)
	}

	aliasMaxDisplayLength := columnWidthAlias - (paddingHorizontal * 2)

	if len(alias) > aliasMaxDisplayLength {
		alias = alias[:aliasMaxDisplayLength-1] + "…"
	}

	return []string{id, alias, doc.GetMetadata().GetVersion(), fmt.Sprint(len(doc.GetNodeList().GetNodes()))}
}
