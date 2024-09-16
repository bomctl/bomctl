// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/export/export.go
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

package export

import (
	"fmt"
	"os"
	"slices"

	"github.com/protobom/protobom/pkg/sbom"
	protowriter "github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func Export(sbomID string, opts *options.ExportOptions) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	opts.Logger.Info("Exporting document", "sbomID", sbomID)

	writer := protowriter.New(protowriter.WithFormat(opts.Format))

	document, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	cleanupEdges(document)

	if opts.OutputFile != nil {
		// Write the SBOM document bytes to file.
		if err := writer.WriteFile(document, opts.OutputFile.Name()); err != nil {
			return fmt.Errorf("%w", err)
		}
	} else {
		// Write the SBOM document bytes to stdout.
		if err := writer.WriteStream(document, os.Stdout); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// With the M2M relationship of NodeLists and Nodes, there needs to be a check/removal of root nodes in other
// docuements so that the current document can serialize correctly.
func cleanupEdges(document *sbom.Document) {
	nodeIDs := []string{}
	for _, node := range document.GetNodeList().GetNodes() {
		nodeIDs = append(nodeIDs, node.GetId())
	}

	edges := []*sbom.Edge{}

	for _, edge := range document.GetNodeList().GetEdges() {
		if slices.Contains(nodeIDs, edge.GetFrom()) {
			edges = append(edges, edge)
		}
	}

	document.NodeList.Edges = edges
}
