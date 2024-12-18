// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/outpututil/table.go
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

package outpututil

import (
	"fmt"
	"os"
	"strconv"
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

	rowHeaderIdx = -1
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
		asList  bool
	}

	TableOption func(*Table)
)

func (t *Table) AddRow(doc *sbom.Document, backend *db.Backend) {
	t.rows = append(t.rows, newRow(doc, backend))
}

func (t *Table) String() string {
	if t.asList || !t.canFit() {
		return t.formatList()
	}

	return t.formatTable()
}

func (t *Table) canFit() bool {
	terminalWidth := termInfo()
	padding := paddingHorizontal * cellSideCount

	for _, row := range t.rows {
		t.columns[columnIdxID].width = max(t.columns[columnIdxID].width, len(row.id)+padding)
		t.columns[columnIdxAlias].width = max(t.columns[columnIdxAlias].width, len(row.alias)+padding)
		t.columns[columnIdxVersion].width = max(t.columns[columnIdxVersion].width, len(row.version)+padding)
		t.columns[columnIdxNumNodes].width = max(t.columns[columnIdxNumNodes].width, len(row.numNodes)+padding)
	}

	return t.getTableWidth() < terminalWidth
}

func (t *Table) formatList() string {
	renderedRows := []string{}

	for _, row := range t.rows {
		renderedRows = append(renderedRows, strings.Join([]string{
			fmt.Sprintf("%-8s: %s", columnNameID, row.id),
			fmt.Sprintf("%-8s: %s", columnNameAlias, row.alias),
			fmt.Sprintf("%-8s: %s", columnNameVersion, row.version),
			fmt.Sprintf("%-8s: %s", columnNameNumNodes, row.numNodes),
		}, "\n"))
	}

	return strings.Join(renderedRows, "\n\n")
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

func NewTable(opts ...TableOption) *Table {
	table := &Table{}

	for idx := range totalColumnCount {
		var column columnDefinition

		switch idx {
		case columnIdxID:
			column.name, column.width = columnNameID, columnWidthID
		case columnIdxAlias:
			column.name, column.width = columnNameAlias, columnWidthAlias
		case columnIdxVersion:
			column.name, column.width = columnNameVersion, columnWidthVersion
		case columnIdxNumNodes:
			column.name, column.width = columnNameNumNodes, columnWidthNumNodes
		}

		table.columns = append(table.columns, column)
	}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func newRow(doc *sbom.Document, backend *db.Backend) rowData {
	id := doc.GetMetadata().GetId()

	alias, err := backend.GetDocumentUniqueAnnotation(id, db.AliasAnnotation)
	if err != nil {
		backend.Logger.Fatalf("failed to get alias: %v", err)
	}

	return rowData{
		id:       id,
		alias:    alias,
		version:  doc.GetMetadata().GetVersion(),
		numNodes: strconv.Itoa(len(doc.GetNodeList().GetNodes())),
	}
}

func termInfo() int {
	fd := int(os.Stdout.Fd())

	width, _, err := term.GetSize(fd)
	if err != nil {
		return 0
	}

	return width
}

func WithListFormat() TableOption {
	return func(t *Table) {
		t.asList = true
	}
}
