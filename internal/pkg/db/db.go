// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/db/db.go
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
package db

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/logger"
)

const (
	AliasAnnotation string = "bomctl_annotation_alias"
	TagAnnotation   string = "bomctl_annotation_tag"

	DatabaseFile string = "bomctl.db"

	EntDebugLevel int = 2
)

type (
	Backend struct {
		*ent.Backend
		*log.Logger
		Verbosity int
	}

	BackendKey struct{}

	Option func(*Backend)
)

var (
	errBackendMissingFromContext = errors.New("failed to get database backend from command context")
	errMultipleDocuments         = errors.New("multiple documents matching ID")
	errInvalidAlias              = errors.New("invalid alias provided")
	errAliasAlreadyExists        = errors.New("alias already exists")
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
		opt(backend)
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

// AddDocument adds the protobom Document to the database.
func (backend *Backend) AddDocument(document *sbom.Document) error {
	if err := backend.Store(document, nil); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
}

// GetDocumentByID retrieves a protobom Document with the specified ID from the database.
func (backend *Backend) GetDocumentByID(id string) (doc *sbom.Document, err error) { //nolint:varnamelen
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

func (backend *Backend) GetDocumentByIDOrAlias(id string) (*sbom.Document, error) { //nolint:varnamelen
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

func (backend *Backend) FilterDocumentsByTag(documents []*sbom.Document, tags ...string) ([]*sbom.Document, error) {
	taggedDocuments, err := backend.GetDocumentsByAnnotation(TagAnnotation, tags...)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by tag: %w", err)
	}

	taggedDocumentIDs := make([]string, len(taggedDocuments))
	for idx := range taggedDocuments {
		taggedDocumentIDs[idx] = taggedDocuments[idx].GetMetadata().GetId()
	}

	filteredDocuments := []*sbom.Document{}

	for _, doc := range documents {
		if slices.Contains(taggedDocumentIDs, doc.GetMetadata().GetId()) {
			filteredDocuments = append(filteredDocuments, doc)
		}
	}

	documents = filteredDocuments

	return documents, nil
}

func (backend *Backend) SetAlias(documentID, alias string) (err error) {
	if err := backend.validateNewAlias(alias); err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	docAlias, err := backend.GetDocumentUniqueAnnotation(documentID, AliasAnnotation)
	if err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	if err := backend.RemoveAnnotations(documentID, AliasAnnotation, docAlias); err != nil {
		return fmt.Errorf("failed to remove previous alias: %w", err)
	}

	if err := backend.SetUniqueAnnotation(documentID, AliasAnnotation, alias); err != nil {
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
		err = errAliasAlreadyExists
	}

	return err
}

// WithDatabaseFile sets the database file for the backend.
func WithDatabaseFile(file string) Option {
	return func(backend *Backend) {
		backend.Backend.Options.DatabaseFile = file
	}
}

// WithVerbosity sets the SQL debugging level for the backend.
func WithVerbosity(verbosity int) Option {
	return func(backend *Backend) {
		backend.Verbosity = verbosity
	}
}
