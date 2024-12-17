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
	"strings"

	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

func AddLink(backend *db.Backend, opts *options.LinkOptions) error {
	link := opts.Links[0]

	switch link.From.Type {
	case options.LinkTargetTypeDocument:
		if err := backend.AddDocumentAnnotations(link.From.ID, db.LinkToAnnotation, link.To[0].ID); err != nil {
			return fmt.Errorf("%w", err)
		}

		opts.Logger.Info("Added document links", "from", link.From.String(), "to", link.To[0].String())
	case options.LinkTargetTypeNode:
		if err := backend.AddNodeAnnotations(link.From.ID, db.LinkToAnnotation, link.To[0].ID); err != nil {
			return fmt.Errorf("adding node link: %w", err)
		}

		opts.Logger.Info("Added node links", "from", link.From.String(), "to", link.To[0].String())
	}

	return nil
}

func ClearLinks(backend *db.Backend, opts *options.LinkOptions) error {
	for _, link := range opts.Links {
		switch link.From.Type {
		case options.LinkTargetTypeDocument:
			annotations, err := backend.GetDocumentAnnotations(link.From.ID, db.LinkToAnnotation)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			if err := backend.RemoveDocumentAnnotations(link.From.ID, db.LinkToAnnotation); err != nil {
				return fmt.Errorf("%w", err)
			}

			cleared := sliceutil.Extract(annotations, func(a *ent.Annotation) string { return a.Value })

			opts.Logger.Info("Cleared document links", "from", link.From.String(), "to", cleared)
		case options.LinkTargetTypeNode:
			annotations, err := backend.GetNodeAnnotations(link.From.ID, db.LinkToAnnotation)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			if err := backend.RemoveNodeAnnotations(link.From.ID, db.LinkToAnnotation); err != nil {
				return fmt.Errorf("%w", err)
			}

			cleared := sliceutil.Extract(annotations, func(a *ent.Annotation) string { return a.Value })

			opts.Logger.Info("Cleared node links", "from", link.From.String(), "to", cleared)
		}
	}

	return nil
}

func ListLinks(backend *db.Backend, opts *options.LinkOptions) ([]options.LinkTarget, error) {
	var (
		annotations ent.Annotations
		documents   []*sbom.Document
		nodes       []*sbom.Node
		err         error
	)

	link := opts.Links[0]

	switch link.From.Type {
	case options.LinkTargetTypeDocument:
		// Get outgoing links for the document.
		annotations, err = backend.GetDocumentAnnotations(link.From.ID, db.LinkToAnnotation)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		// Get incoming links for the document.
		documents, err = backend.GetDocumentsByAnnotation(db.LinkToAnnotation, link.From.ID)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		nodes, err = backend.GetNodesByAnnotation(db.LinkToAnnotation, link.From.ID)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	case options.LinkTargetTypeNode:
		// Get outgoing links for the node.
		annotations, err = backend.GetNodeAnnotations(link.From.ID, db.LinkToAnnotation)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	for _, annotation := range annotations {
		alias := backend.GetDocumentAlias(annotation.Value)
		link.To = append(link.To, options.LinkTarget{
			ID:    annotation.Value,
			Alias: alias,
			Type:  options.LinkTargetTypeDocument,
		})
	}

	// Update value of provided links.
	opts.Links = []options.Link{link}

	incoming := sliceutil.Extract(documents, func(d *sbom.Document) options.LinkTarget {
		return options.LinkTarget{
			ID:    d.GetMetadata().GetId(),
			Alias: backend.GetDocumentAlias(d.GetMetadata().GetId()),
			Type:  options.LinkTargetTypeDocument,
		}
	})

	incoming = append(incoming, sliceutil.Extract(nodes, func(n *sbom.Node) options.LinkTarget {
		return options.LinkTarget{
			ID:    n.GetId(),
			Alias: backend.GetDocumentAlias(n.GetId()),
			Type:  options.LinkTargetTypeDocument,
		}
	})...)

	return incoming, nil
}

func RemoveLink(backend *db.Backend, opts *options.LinkOptions) (err error) {
	for _, link := range opts.Links {
		toIDs := sliceutil.Extract(link.To, func(lt options.LinkTarget) string { return lt.ID })

		switch link.From.Type {
		case options.LinkTargetTypeDocument:
			if err := backend.RemoveDocumentAnnotations(link.From.ID, db.LinkToAnnotation, toIDs...); err != nil {
				return fmt.Errorf("%w", err)
			}

			opts.Logger.Info("Removed document links", "from", link.From.String(), "to", strings.Join(toIDs, "\n\t"))
		case options.LinkTargetTypeNode:
			if err := backend.RemoveNodeAnnotations(link.From.ID, db.LinkToAnnotation, toIDs...); err != nil {
				return fmt.Errorf("%w", err)
			}

			opts.Logger.Info("Removed node links", "from", link.From.String(), "to", strings.Join(toIDs, "\n\t"))
		}
	}

	return nil
}
