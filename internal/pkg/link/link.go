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
	if len(opts.Links) != 1 || len(opts.Links[0].To) != 1 {
		opts.Logger.Warn("Expected 1 link source", "got", len(opts.Links))

		if len(opts.Links) != 0 {
			opts.Logger.Warn("Expected 1 link target", "got", len(opts.Links[0].To))
		}

		return nil
	}

	source, target := opts.Links[0].From, opts.Links[0].To[0]

	switch source.Type {
	case options.LinkTargetTypeDocument:
		if err := backend.AddDocumentAnnotations(source.ID, db.LinkToAnnotation, target.ID); err != nil {
			return fmt.Errorf("%w", err)
		}

		opts.Logger.Info("Added document link", "from", source.String(), "to", target.String())
	case options.LinkTargetTypeNode:
		if err := backend.AddNodeAnnotations(source.ID, db.LinkToAnnotation, target.ID); err != nil {
			return fmt.Errorf("adding node link: %w", err)
		}

		opts.Logger.Info("Added node link", "from", source.String(), "to", target.String())
	}

	return nil
}

func ClearLinks(backend *db.Backend, opts *options.LinkOptions) error {
	var (
		annotations ent.Annotations
		err         error
	)

	for _, link := range opts.Links {
		switch link.From.Type {
		case options.LinkTargetTypeDocument:
			annotations, err = backend.GetDocumentAnnotations(link.From.ID, db.LinkToAnnotation)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			if err := backend.RemoveDocumentAnnotations(link.From.ID, db.LinkToAnnotation); err != nil {
				return fmt.Errorf("%w", err)
			}
		case options.LinkTargetTypeNode:
			annotations, err = backend.GetNodeAnnotations(link.From.ID, db.LinkToAnnotation)
			if err != nil {
				return fmt.Errorf("%w", err)
			}

			if err := backend.RemoveNodeAnnotations(link.From.ID, db.LinkToAnnotation); err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		link.To = append(link.To, sliceutil.Extract(annotations, func(a *ent.Annotation) options.LinkTarget {
			return options.LinkTarget{
				ID:    a.Value,
				Alias: backend.GetDocumentAlias(a.Value),
				Type:  options.LinkTargetTypeDocument,
			}
		})...)

		cleared := strings.Join(
			sliceutil.Extract(link.To, func(lt options.LinkTarget) string { return lt.String() }),
			", ",
		)

		opts.Logger.Info(fmt.Sprintf("Cleared %s links", link.From.Type), "from", link.From.String(), "to", cleared)
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
		link.To = append(link.To, options.LinkTarget{
			ID:    annotation.Value,
			Alias: backend.GetDocumentAlias(annotation.Value),
			Type:  options.LinkTargetTypeDocument,
		})
	}

	// Update value of provided links.
	opts.Links = []options.Link{link}

	incoming := sliceutil.Extract(documents, func(d *sbom.Document) options.LinkTarget {
		id := d.GetMetadata().GetId()

		return options.LinkTarget{
			ID:    id,
			Alias: backend.GetDocumentAlias(id),
			Type:  options.LinkTargetTypeDocument,
		}
	})

	incoming = append(incoming, sliceutil.Extract(nodes, func(n *sbom.Node) options.LinkTarget {
		return options.LinkTarget{
			ID:   n.GetId(),
			Type: options.LinkTargetTypeNode,
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
