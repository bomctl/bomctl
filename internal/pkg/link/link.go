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
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const cyan = lipgloss.ANSIColor(38)

func AddLink(backend *db.Backend, opts *options.LinkOptions) error {
	var err error

	switch {
	case len(opts.DocumentIDs) > 0:
		err = backend.AddDocumentAnnotations(opts.DocumentIDs[0], db.LinkToAnnotation, opts.ToIDs[0])
	case len(opts.NodeIDs) > 0:
		err = backend.AddNodeAnnotations(opts.NodeIDs[0], db.LinkToAnnotation, opts.ToIDs[0])
	}

	if err != nil {
		return fmt.Errorf("%w", err)
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
		annotations   ent.Annotations
		incomingLinks []string
		id, fromType  string
		err           error
	)

	switch {
	case len(opts.DocumentIDs) > 0:
		fromType = "document"
		id = opts.DocumentIDs[0]
		annotations, err = backend.GetDocumentAnnotations(id, db.LinkToAnnotation)

		// Get incoming links for the docucment.
		documents, err := backend.GetDocumentsByAnnotation(db.LinkToAnnotation, opts.DocumentIDs...)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		incomingLinks = sliceutil.Extract(documents, func(d *sbom.Document) string {
			return d.GetMetadata().GetId()
		})
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

	children := sliceutil.Extract(annotations, func(annotation *ent.Annotation) string {
		return annotation.Value
	})

	style := lipgloss.NewStyle().Foreground(cyan)
	links := tree.Root(fmt.Sprintf("Links for %s %s:", fromType, style.Render(id))).
		Child(children).
		EnumeratorStyle(style)

	fmt.Fprintln(os.Stdout, links)

	if len(incomingLinks) > 0 {
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, tree.Root("Incoming links:").
			Child(incomingLinks).
			EnumeratorStyle(style),
		)
	}

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
