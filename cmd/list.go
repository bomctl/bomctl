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

			rows := [][]string{}
			for _, document := range documents {
				rows = append(rows, getRow(document))
			}

			fmt.Fprintf(os.Stdout, "\n%s\n\n", table.New().
				Headers("ID", "Version", "# Nodes").
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

func getRow(doc *sbom.Document) []string {
	id := doc.GetMetadata().GetName()
	if id == "" {
		id = doc.GetMetadata().GetId()
	}

	return []string{id, doc.GetMetadata().GetVersion(), fmt.Sprint(len(doc.GetNodeList().GetNodes()))}
}
