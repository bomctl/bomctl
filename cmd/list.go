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

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/spf13/cobra"
	"golang.org/x/term"

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

	cellSideCount = 2

	headerCount = 4

	paddingHorizontal = 1
	paddingVertical   = 0

	rowHeaderIdx = 0
	rowMaxHeight = 1
)

func getTermInfo() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		return 0
	}

	return width
}

func determineColumnWidths(rows ...[]string) (tooWide bool, colWidths []int) {
	terminalWidth := getTermInfo()
	idColumnWidth := columnWidthID
	aliasColumnWidth := columnWidthAlias

	for _, row := range rows {
		if idWidth := len(row[columnIdxID]); idWidth >= idColumnWidth {
			idColumnWidth = idWidth + (paddingHorizontal * cellSideCount)
		}

		if aliasWidth := len(row[columnIdxAlias]); aliasWidth >= aliasColumnWidth {
			aliasColumnWidth = aliasWidth + (paddingHorizontal * cellSideCount)
		}
	}

	totalWidth := (idColumnWidth + aliasColumnWidth + columnWidthVersion + columnWidthNumNodes)
	totalWidth += headerCount * (paddingHorizontal * cellSideCount)

	return (totalWidth > terminalWidth), []int{idColumnWidth, aliasColumnWidth, columnWidthVersion, columnWidthNumNodes}
}

func printList(rows ...[]string) {
	for _, row := range rows {
		fmt.Fprintf(os.Stdout, "ID: %s\n"+"Alias: %s\n"+"Version: %s\n"+"# Nodes: %s\n\n",
			row[columnIdxID], row[columnIdxAlias], row[columnIdxVersion], row[columnIdxNumNodes])
	}
}

func printTable(colWidths []int, rows ...[]string) {
	commonStyle := lipgloss.NewStyle().
		Padding(paddingVertical, paddingHorizontal).
		MaxHeight(rowMaxHeight)

	fmt.Fprintf(os.Stdout, "\n%s\n\n", table.New().
		Headers("ID", "Alias", "Version", "# Nodes").
		Rows(rows...).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			align := lipgloss.Center
			width := 0

			switch col {
			case columnIdxID:
				width = colWidths[columnIdxID]

				if row != rowHeaderIdx {
					align = lipgloss.Left
				}
			case columnIdxAlias:
				width = colWidths[columnIdxAlias]
			case columnIdxVersion:
				width = colWidths[columnIdxVersion]
			case columnIdxNumNodes:
				width = colWidths[columnIdxNumNodes]
			}

			return commonStyle.Width(width).AlignHorizontal(align)
		}).
		Render(),
	)
}

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

			tooWide, columnWidths := determineColumnWidths(rows...)

			if tooWide {
				printList(rows...)
			} else {
				printTable(columnWidths, rows...)
			}
		},
	}

	listCmd.Flags().StringArrayVar(&tags, "tag", []string{},
		"Tag(s) used to filter documents (can be specified multiple times)")

	return listCmd
}

func getRow(doc *sbom.Document, backend *db.Backend) []string {
	id := doc.GetMetadata().GetId()

	alias, err := backend.GetDocumentUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation)
	if err != nil {
		backend.Logger.Fatalf("failed to get alias: %v", err)
	}

	return []string{id, alias, doc.GetMetadata().GetVersion(), fmt.Sprint(len(doc.GetNodeList().GetNodes()))}
}
