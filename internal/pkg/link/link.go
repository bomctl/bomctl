// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/link/link.go
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

package link

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/outpututil"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const cyan = lipgloss.ANSIColor(38)

func AddLink(backend *db.Backend, opts *options.LinkOptions) error {
	toID, err := resolveDocumentID(opts.ToIDs[0], backend)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	switch {
	case len(opts.DocumentIDs) > 0:
		opts.Logger.Info("Adding document link", "from", opts.DocumentIDs[0], "to", toID)

		fromID, err := resolveDocumentID(opts.DocumentIDs[0], backend)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if err := backend.AddDocumentAnnotations(fromID, db.LinkToAnnotation, toID); err != nil {
			return fmt.Errorf("%w", err)
		}
	case len(opts.NodeIDs) > 0:
		opts.Logger.Info("Adding node link", "from", opts.NodeIDs[0], "to", toID)

		if err := backend.AddNodeAnnotations(opts.NodeIDs[0], db.LinkToAnnotation, toID); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func ClearLinks(backend *db.Backend, opts *options.LinkOptions) error {
	for _, from := range opts.DocumentIDs {
		if err := backend.RemoveDocumentAnnotations(from, db.LinkToAnnotation); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	for _, from := range opts.NodeIDs {
		if err := backend.RemoveNodeAnnotations(from, db.LinkToAnnotation); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func ListLinks(backend *db.Backend, opts *options.LinkOptions) error {
	var (
		annotations  ent.Annotations
		documents    []*sbom.Document
		id, fromType string
		err          error
	)

	switch {
	case len(opts.DocumentIDs) > 0:
		fromType = "document"
		id = opts.DocumentIDs[0]

		var nativeID string

		nativeID, err = resolveDocumentID(id, backend)
		if err != nil {
			break
		}

		annotations, err = backend.GetDocumentAnnotations(nativeID, db.LinkToAnnotation)
		if err != nil {
			break
		}

		// Get incoming links for the document.
		documents, err = backend.GetDocumentsByAnnotation(db.LinkToAnnotation, nativeID)
		if err != nil {
			break
		}
	case len(opts.NodeIDs) > 0:
		fromType = "node"
		id = opts.NodeIDs[0]
		annotations, err = backend.GetNodeAnnotations(id, db.LinkToAnnotation)
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(annotations) == 0 {
		opts.Logger.Warn("No links found for "+fromType, "id", id)

		return nil
	}

	linkIDs := sliceutil.Extract(annotations, func(annotation *ent.Annotation) string {
		return annotation.Value
	})

	writeLinksTree(id, fromType, linkIDs, documents, backend)

	return nil
}

func RemoveLink(backend *db.Backend, opts *options.LinkOptions) error {
	var err error

	switch {
	case len(opts.DocumentIDs) > 0:
		err = backend.RemoveDocumentAnnotations(opts.DocumentIDs[0], db.LinkToAnnotation, opts.ToIDs...)
	case len(opts.NodeIDs) > 0:
		err = backend.RemoveNodeAnnotations(opts.NodeIDs[0], db.LinkToAnnotation, opts.ToIDs...)
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func resolveDocumentID(id string, backend *db.Backend) (string, error) {
	document, err := backend.GetDocumentByIDOrAlias(id)
	if err != nil {
		return "", fmt.Errorf("resolving document ID: %w", err)
	}

	if document == nil {
		backend.Logger.Warn("Document not found", "id", id)

		return "", nil
	}

	return document.GetMetadata().GetId(), nil
}

func writeLinksTree(id, fromType string, linkIDs []string, documents []*sbom.Document, backend *db.Backend) {
	style := lipgloss.NewStyle().Foreground(cyan)
	links := tree.Root(fmt.Sprintf("Links for %s %s:", fromType, style.Render(id))).
		Child(linkIDs).
		EnumeratorStyle(style)

	fmt.Fprintln(os.Stdout, links)

	if len(documents) > 0 {
		incomingLinks := outpututil.NewTable(outpututil.WithListFormat())
		for _, document := range documents {
			incomingLinks.AddRow(document, backend)
		}

		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, tree.Root("Incoming links:").
			Child(incomingLinks).
			Enumerator(func(_ tree.Children, _ int) string { return "\t" }),
		)
	}
}
