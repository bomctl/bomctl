// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/format/table.go
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

package outpututils

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	lgtable "github.com/charmbracelet/lipgloss/table"
	"github.com/protobom/protobom/pkg/sbom"
	"golang.org/x/term"

	"github.com/bomctl/bomctl/internal/pkg/db"
)

const (
	columnIdxID = iota
	columnIdxAlias
	columnIdxVersion
	columnIdxNumNodes

	columnNameID       = "ID"
	columnNameAlias    = "Alias"
	columnNameVersion  = "Version"
	columnNameNumNodes = "# Nodes"

	columnWidthID       = 47
	columnWidthAlias    = 12
	columnWidthVersion  = 9
	columnWidthNumNodes = 9

	cellSideCount = 2

	paddingHorizontal = 1
	paddingVertical   = 0

	rowHeaderIdx = 0
	rowMaxHeight = 1

	totalColumnCount = 4
)

type (
	columnDefinition struct {
		name  string
		width int
	}

	rowData struct {
		id, alias, version, numNodes string
	}

	Table struct {
		columns []columnDefinition
		rows    []rowData
	}
)

func (t *Table) String() string {
	tooWide := t.determineColumnWidths()

	if tooWide {
		return t.formatList()
	}

	return t.formatTable()
}

func NewTable() *Table {
	cols := []columnDefinition{}

	for c := range totalColumnCount {
		name := ""
		width := 0

		switch c {
		case columnIdxID:
			name = columnNameID
			width = columnWidthID
		case columnIdxAlias:
			name = columnNameAlias
			width = columnWidthAlias
		case columnIdxVersion:
			name = columnNameVersion
			width = columnWidthVersion
		case columnIdxNumNodes:
			name = columnNameNumNodes
			width = columnWidthNumNodes
		}

		cols = append(cols, columnDefinition{
			name:  name,
			width: width,
		})
	}

	return &Table{
		columns: cols,
		rows:    []rowData{},
	}
}

func (t *Table) AddRow(doc *sbom.Document, backend *db.Backend) {
	t.rows = append(t.rows, getRow(doc, backend))
}

func (t *Table) getHeaders() []string {
	headers := []string{}

	for c := range totalColumnCount {
		headers = append(headers, t.columns[c].name)
	}

	return headers
}

func (t *Table) getRows() [][]string {
	rows := [][]string{}

	for _, row := range t.rows {
		rows = append(rows, []string{row.id, row.alias, row.version, row.numNodes})
	}

	return rows
}

func (t *Table) getTableWidth() int {
	totalWidth := totalColumnCount * (paddingHorizontal * cellSideCount)

	for c := range totalColumnCount {
		totalWidth += t.columns[c].width
	}

	return totalWidth
}

func (t *Table) determineColumnWidths() bool {
	terminalWidth := getTermInfo()
	padding := paddingHorizontal * cellSideCount

	for _, row := range t.rows {
		if idWidth := len(row.id); idWidth >= t.columns[columnIdxID].width {
			t.columns[columnIdxID].width = idWidth + padding
		}

		if aliasWidth := len(row.alias); aliasWidth >= t.columns[columnIdxAlias].width {
			t.columns[columnIdxAlias].width = aliasWidth + padding
		}

		if versionWidth := len(row.version); versionWidth >= columnIdxVersion {
			t.columns[columnIdxVersion].width = versionWidth + padding
		}

		if numNodeWidth := len(row.numNodes); numNodeWidth >= t.columns[columnIdxNumNodes].width {
			t.columns[columnIdxNumNodes].width = numNodeWidth + padding
		}
	}

	return (t.getTableWidth() > terminalWidth)
}

func (t *Table) formatList() string {
	output := ""
	for _, row := range t.rows {
		output = strings.Join([]string{
			output,
			fmt.Sprintf("%-8s: %s", columnNameID, row.id),
			fmt.Sprintf("%-8s: %s", columnNameAlias, row.alias),
			fmt.Sprintf("%-8s: %s", columnNameVersion, row.version),
			fmt.Sprintf("%-8s: %s", columnNameNumNodes, row.numNodes),
			"",
		}, "\n")
	}

	return output
}

func (t *Table) formatTable() string {
	commonStyle := lipgloss.NewStyle().
		Padding(paddingVertical, paddingHorizontal).
		MaxHeight(rowMaxHeight)

	return lgtable.New().
		Headers(t.getHeaders()...).
		Rows(t.getRows()...).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(true).
		StyleFunc(func(row, col int) lipgloss.Style {
			align := lipgloss.Center
			width := 0
			columns := t.columns

			switch col {
			case columnIdxID:
				width = columns[columnIdxID].width

				if row != rowHeaderIdx {
					align = lipgloss.Left
				}
			case columnIdxAlias:
				width = columns[columnIdxAlias].width
			case columnIdxVersion:
				width = columns[columnIdxVersion].width
			case columnIdxNumNodes:
				width = columns[columnIdxNumNodes].width
			}

			return commonStyle.Width(width).AlignHorizontal(align)
		}).
		Render()
}

func getTermInfo() int {
	width, _, err := term.GetSize(0)
	if err != nil {
		return 0
	}

	return width
}

func getRow(doc *sbom.Document, backend *db.Backend) rowData {
	id := doc.GetMetadata().GetId()

	alias, err := backend.GetDocumentUniqueAnnotation(doc.GetMetadata().GetId(), db.AliasAnnotation)
	if err != nil {
		backend.Logger.Fatalf("failed to get alias: %v", err)
	}

	return rowData{
		id:       id,
		alias:    alias,
		version:  doc.GetMetadata().GetVersion(),
		numNodes: fmt.Sprint(len(doc.GetNodeList().GetNodes())),
	}
}
