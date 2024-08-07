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
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/utils"
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
	documentIDs := []string{}

	listCmd := &cobra.Command{
		Use:     "list [flags] SBOM_ID...",
		Aliases: []string{"ls"},
		Short:   "List SBOM documents in local cache",
		Long:    "List SBOM documents in local cache",
		PreRun: func(_ *cobra.Command, args []string) {
			documentIDs = append(documentIDs, args...)
		},
		Run: func(cmd *cobra.Command, _ []string) {
			verbosity, err := cmd.Flags().GetCount("verbose")
			cobra.CheckErr(err)

			backend := db.NewBackend().
				Debug(verbosity >= minDebugLevel).
				WithDatabaseFile(filepath.Join(viper.GetString("cache_dir"), db.DatabaseFile)).
				WithLogger(utils.NewLogger("list"))

			if err := backend.InitClient(); err != nil {
				backend.Logger.Fatalf("failed to initialize backend client: %v", err)
			}

			defer backend.CloseClient()

			documents, err := backend.GetDocumentsByID(documentIDs...)
			if err != nil {
				backend.Logger.Fatalf("failed to get documents: %v", err)
			}

			rows := [][]string{}
			for _, document := range documents {
				id := document.Metadata.Name
				if id == "" {
					id = document.Metadata.Id
				}

				rows = append(rows, []string{
					id, document.Metadata.Version, fmt.Sprint(len(document.NodeList.Nodes)),
				})
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
