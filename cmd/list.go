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
	columnIdxVersion
	columnIdxNumNodes

	columnWidthID       = 50
	columnWidthVersion  = 10
	columnWidthNumNodes = 10

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

			documents, err := backend.GetDocumentsByID(args...)
			if err != nil {
				backend.Logger.Fatalf("failed to get documents: %v", err)
			}

			documents, err = backend.SelectDocumentsByTag(documents, tags...)
			if err != nil {
				backend.Logger.Fatalf("failed to get documents: %v", err)
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
	id := doc.Metadata.Name
	if id == "" {
		id = doc.Metadata.Id
	}

	alias, err := backend.GetDocumentAlias(doc.Metadata.Id)
	if err != nil {
		backend.Logger.Fatalf("failed to get alias: %v", err)
	}

	return []string{id, alias, doc.Metadata.Version, fmt.Sprint(len(doc.NodeList.Nodes))}
}
