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
	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const cyan = lipgloss.ANSIColor(38)

type FromType interface {
	*sbom.Document | *sbom.Node
}

func AddLink[T FromType](from T, documentID string, backend *db.Backend) error {
	document, err := backend.GetDocumentByIDOrAlias(documentID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	documentUUID, err := ent.GenerateUUID(document)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	switch from := any(from).(type) {
	case *sbom.Document:
		err = backend.AddDocumentAnnotations(from.GetMetadata().GetId(), db.LinkToAnnotation, documentUUID.String())
	case *sbom.Node:
		err = backend.AddNodeAnnotations(from.GetId(), db.LinkToAnnotation, documentUUID.String())
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func ClearLinks[T FromType](from T, backend *db.Backend) error {
	var err error

	switch from := any(from).(type) {
	case *sbom.Document:
		err = backend.RemoveDocumentAnnotations(from.GetMetadata().GetId(), db.LinkToAnnotation)
	case *sbom.Node:
		err = backend.RemoveNodeAnnotations(from.GetId(), db.LinkToAnnotation)
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func ListLinks[T FromType](from T, backend *db.Backend) error {
	var (
		annotations  ent.Annotations
		id, fromType string
		err          error
	)

	switch from := any(from).(type) {
	case *sbom.Document:
		fromType = "document"
		id = from.GetMetadata().GetId()
		annotations, err = backend.GetDocumentAnnotations(id, db.LinkToAnnotation)
	case *sbom.Node:
		fromType = "node"
		id = from.GetId()
		annotations, err = backend.GetNodeAnnotations(id, db.LinkToAnnotation)
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(annotations) == 0 {
		log.Warnf("No links found for %s: %s", fromType, id)

		return nil
	}

	children := sliceutil.Extract(annotations, func(annotation *ent.Annotation) string {
		return annotation.DocumentID.String()
	})

	style := lipgloss.NewStyle().Foreground(cyan)
	links := tree.Root(fmt.Sprintf("Links for %s:", style.Render(id))).
		Child(children).
		EnumeratorStyle(style)

	fmt.Fprintln(os.Stdout, links)

	return nil
}

func RemoveLink[T FromType](from T, documentID string, backend *db.Backend) error {
	document, err := backend.GetDocumentByIDOrAlias(documentID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	documentUUID, err := ent.GenerateUUID(document)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	switch from := any(from).(type) {
	case *sbom.Document:
		id := from.GetMetadata().GetId()
		err = backend.RemoveDocumentAnnotations(id, db.LinkToAnnotation, documentUUID.String())
	case *sbom.Node:
		id := from.GetId()
		err = backend.RemoveNodeAnnotations(id, db.LinkToAnnotation, documentUUID.String())
	}

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
