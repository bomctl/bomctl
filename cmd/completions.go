// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/completions.go
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
	"path/filepath"
	"slices"
	"strings"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

func completions(
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	cacheDir := viper.GetString("cache_dir")

	backend, err := db.NewBackend(db.WithDatabaseFile(filepath.Join(cacheDir, db.DatabaseFile)))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	defer backend.CloseClient()

	comps := []string{}

	documents, err := backend.GetDocumentsByIDOrAlias()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	documentIDs := sliceutil.Extract(documents, func(doc *sbom.Document) string {
		return doc.GetMetadata().GetId()
	})

	for _, documentID := range documentIDs {
		if slices.Contains(comps, documentID) {
			continue
		}

		if strings.HasPrefix(documentID, toComplete) {
			comps = cobra.AppendActiveHelp(comps, documentID)

			continue
		}

		annotations, err := backend.GetDocumentAnnotations(documentID)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		annotation, err := sliceutil.Next(annotations, func(a *ent.Annotation) bool {
			return strings.HasPrefix(a.Name, db.AliasAnnotation) && strings.HasPrefix(a.Value, toComplete)
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("%s (%s)", documentID, annotation.Value))
	}

	return comps, cobra.ShellCompDirectiveNoFileComp
}
