// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/db/db.go
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

package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/storage"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/logger"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

const (
	AliasAnnotation           string = "bomctl_annotation_alias"
	BaseDocumentAnnotation    string = "bomctl_annotation_base_document"
	RevisedDocumentAnnotation string = "bomctl_annotation_revised_document"
	LatestRevisionAnnotation  string = "bomctl_annotation_latest_revision"
	SourceDataAnnotation      string = "bomctl_annotation_source_data"
	SourceFormatAnnotation    string = "bomctl_annotation_source_format"
	SourceHashAnnotation      string = "bomctl_annotation_source_hash"
	SourceURLAnnotation       string = "bomctl_annotation_source_url"
	TagAnnotation             string = "bomctl_annotation_tag"

	DatabaseFile string = "bomctl.db"

	EntDebugLevel int = 2

	OriginalFormat = "original"
)

type (
	Backend struct {
		*ent.Backend
		*log.Logger
		Verbosity int
	}

	BackendKey struct{}

	Option func(*Backend) error
)

var (
	errBackendMissingFromContext = errors.New("failed to get database backend from command context")
	errMultipleDocuments         = errors.New("multiple documents matching ID")
	errInvalidAlias              = errors.New("invalid alias provided")
	errDuplicateAlias            = errors.New("alias already exists")
	ErrDocumentAliasExists       = errors.New("the document already has an alias")
)

func BackendFromContext(ctx context.Context) (*Backend, error) {
	backend, ok := ctx.Value(BackendKey{}).(*Backend)
	if !ok {
		return nil, errBackendMissingFromContext
	}

	return backend, nil
}

func NewBackend(opts ...Option) (*Backend, error) {
	backend := &Backend{Backend: ent.NewBackend(), Logger: logger.New("db")}

	for _, opt := range opts {
		err := opt(backend)
		if err != nil {
			return nil, fmt.Errorf("failed to process backend options: %w", err)
		}
	}

	if backend.Verbosity >= EntDebugLevel {
		backend.Backend.Debug()
	}

	if backend.Backend.Options.DatabaseFile == "" {
		backend.Backend.Options.DatabaseFile = DatabaseFile
	}

	if err := backend.InitClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize backend client: %w", err)
	}

	return backend, nil
}

// AddDocument adds the protobom document to the database and applies given backendOpts.
func (backend *Backend) AddDocument(sbomData []byte, backendOpts ...Option) (*sbom.Document, error) {
	// Clear annotations after adding document.
	defer func() {
		backend.Options.Annotations = nil
	}()

	sbomReader := reader.New()

	document, err := sbomReader.ParseStream(bytes.NewReader(sbomData))
	if err != nil {
		return nil, fmt.Errorf("parsing SBOM data: %w", err)
	}

	// Collect backend options by calling associated functions.
	for _, fn := range backendOpts {
		err := fn(backend)
		if err != nil {
			return nil, fmt.Errorf("handling document annotations: %w", err)
		}
	}

	// Create StoreOptions with the populated backend options struct.
	opts := &storage.StoreOptions{
		BackendOptions: backend.Options,
	}

	if err := backend.Store(document, opts); err != nil {
		return nil, fmt.Errorf("storing document %s: %w", document.GetMetadata().GetId(), err)
	}

	return document, nil
}

func (backend *Backend) FilterDocumentsByTag(documents []*sbom.Document, tags ...string) ([]*sbom.Document, error) {
	taggedDocuments, err := backend.GetDocumentsByAnnotation(TagAnnotation, tags...)
	if err != nil {
		return nil, fmt.Errorf("getting documents with tags %v: %w", tags, err)
	}

	taggedDocumentIDs := sliceutil.Extract(taggedDocuments, func(doc *sbom.Document) string {
		return doc.GetMetadata().GetId()
	})

	documents = sliceutil.Filter(documents, func(doc *sbom.Document) bool {
		return slices.Contains(taggedDocumentIDs, doc.GetMetadata().GetId())
	})

	return documents, nil
}

// GetDocumentByID retrieves a protobom Document with the specified ID from the database.
func (backend *Backend) GetDocumentByID(id string) (doc *sbom.Document, err error) {
	switch documents, getDocsErr := backend.GetDocumentsByID(id); {
	case getDocsErr != nil:
		err = fmt.Errorf("querying documents: %w", getDocsErr)
	case len(documents) == 0:
		doc = nil
	case len(documents) > 1:
		err = fmt.Errorf("%w %s", errMultipleDocuments, id)
	default:
		doc = documents[0]
	}

	return doc, err
}

func (backend *Backend) GetDocumentByIDOrAlias(id string) (*sbom.Document, error) {
	document, err := backend.GetDocumentByID(id)
	if err != nil {
		return nil, fmt.Errorf("document could not be retrieved: %w", err)
	}

	if document == nil {
		switch documents, getDocErr := backend.GetDocumentsByAnnotation(AliasAnnotation, id); {
		case getDocErr != nil:
			err = fmt.Errorf("document could not be retrieved: %w", getDocErr)
		case len(documents) == 0:
			document = nil
		case len(documents) > 1:
			err = fmt.Errorf("%w %s", errMultipleDocuments, id)
		default:
			document = documents[0]
		}
	}

	return document, err
}

func (backend *Backend) GetDocumentsByIDOrAlias(ids ...string) ([]*sbom.Document, error) {
	if len(ids) == 0 {
		documents, err := backend.GetDocumentsByID()
		if err != nil {
			return nil, fmt.Errorf("failed to get documents: %w", err)
		}

		return documents, nil
	}

	documents := make([]*sbom.Document, len(ids))

	for idx := range ids {
		document, err := backend.GetDocumentByIDOrAlias(ids[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to get documents by ID: %w", err)
		}

		if document == nil {
			return []*sbom.Document{}, nil
		}

		documents[idx] = document
	}

	return documents, nil
}

func (backend *Backend) GetDocumentTags(id string) ([]string, error) {
	annotations, err := backend.GetDocumentAnnotations(id, TagAnnotation)
	if err != nil {
		return nil, fmt.Errorf("failed to get document tags: %w", err)
	}

	tags := make([]string, len(annotations))
	for idx := range annotations {
		tags[idx] = annotations[idx].Value
	}

	return tags, nil
}

func (backend *Backend) SetAlias(documentID, alias string, force bool) error { //revive:disable:flag-parameter
	if err := backend.validateNewAlias(alias); err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	docAlias, err := backend.GetDocumentUniqueAnnotation(documentID, AliasAnnotation)
	if err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	if docAlias != "" {
		if !force {
			return ErrDocumentAliasExists
		}

		if err := backend.RemoveDocumentAnnotations(documentID, AliasAnnotation, docAlias); err != nil {
			return fmt.Errorf("failed to remove previous alias: %w", err)
		}
	}

	if err := backend.SetDocumentUniqueAnnotation(documentID, AliasAnnotation, alias); err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	return nil
}

func (backend *Backend) validateNewAlias(alias string) (err error) {
	if isEmptyOrWhitespace := regexp.MustCompile(`^\s*$`).MatchString(alias); isEmptyOrWhitespace {
		return errInvalidAlias
	}

	switch documents, getDocErr := backend.GetDocumentsByAnnotation(AliasAnnotation, alias); {
	case getDocErr != nil:
		err = fmt.Errorf("error checking for existing alias: %w", getDocErr)
	case len(documents) == 0:
		err = nil
	default:
		err = errDuplicateAlias
	}

	return err
}

func WithSourceDocumentAnnotations(sbomData []byte) Option {
	return func(backend *Backend) error {
		hash := sha256.Sum256(sbomData)

		backend.Options.Annotations = append(backend.Options.Annotations,
			&ent.Annotation{
				Name:     SourceDataAnnotation,
				Value:    string(sbomData),
				IsUnique: true,
			},
			&ent.Annotation{
				Name:     SourceHashAnnotation,
				Value:    string(hash[:]),
				IsUnique: true,
			},
			&ent.Annotation{
				Name:     LatestRevisionAnnotation,
				Value:    "true",
				IsUnique: false,
			},
		)

		return nil
	}
}

func WithRevisedDocumentAnnotations(base *sbom.Document) Option {
	return func(backend *Backend) error {
		baseID := base.GetMetadata().GetId()

		baseUUID, err := ent.GenerateUUID(base)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		backend.Options.Annotations = append(backend.Options.Annotations,
			&ent.Annotation{
				Name:     BaseDocumentAnnotation,
				Value:    baseUUID.String(),
				IsUnique: true,
			},
			&ent.Annotation{
				Name:     LatestRevisionAnnotation,
				Value:    "true",
				IsUnique: false,
			},
		)

		if err := backend.RemoveDocumentAnnotations(baseID, LatestRevisionAnnotation); err != nil {
			return fmt.Errorf("failed to remove latest annotation: %w", err)
		}

		if err := backend.AddDocumentAnnotations(baseID, RevisedDocumentAnnotation, "true"); err != nil {
			return fmt.Errorf("failed to add revision annotation: %w", err)
		}

		docAlias, err := backend.GetDocumentUniqueAnnotation(baseID, AliasAnnotation)
		if err != nil {
			return fmt.Errorf("failed checking for existing alias: %w", err)
		}

		// If base doc has existing alias, move to revised doc.
		if docAlias != "" {
			// Remove alias from base document.
			if err := backend.RemoveDocumentAnnotations(baseID, AliasAnnotation, docAlias); err != nil {
				return fmt.Errorf("failed to remove existing alias: %w", err)
			}

			// Add AliasAnnotation to revised document
			backend.Options.Annotations = append(backend.Options.Annotations,
				&ent.Annotation{
					Name:     AliasAnnotation,
					Value:    docAlias,
					IsUnique: true,
				},
			)
		}

		return nil
	}
}

// WithDatabaseFile sets the database file for the backend.
func WithDatabaseFile(file string) Option {
	return func(backend *Backend) error {
		backend.Backend.Options.DatabaseFile = file

		return nil
	}
}

// WithVerbosity sets the SQL debugging level for the backend.
func WithVerbosity(verbosity int) Option {
	return func(backend *Backend) error {
		backend.Verbosity = verbosity

		return nil
	}
}
