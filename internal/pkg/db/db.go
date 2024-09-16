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

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/storage"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/logger"
)

const (
	SourceDataAnnotation   string = "bomctl_annotation_source_data"
	SourceHashAnnotation   string = "bomctl_annotation_source_hash"
	SourceFormatAnnotation string = "bomctl_annotation_source_format"
	SourceURLAnnotation    string = "bomctl_annotation_source_url"

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

var errBackendMissingFromContext = errors.New("failed to get database backend from command context")

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

// AddDocument adds the protobom Document to the database and annotates with its source data, hash, and format.
func (backend *Backend) AddDocument(sbomData []byte) (*sbom.Document, error) {
	sbomReader := reader.New()

	document, err := sbomReader.ParseStream(bytes.NewReader(sbomData))
	if err != nil {
		return nil, fmt.Errorf("parsing SBOM data: %w", err)
	}

	hash := sha256.Sum256(sbomData)
	opts := &storage.StoreOptions{
		BackendOptions: &ent.BackendOptions{
			Annotations: ent.Annotations{
				{
					Name:     SourceDataAnnotation,
					Value:    string(sbomData),
					IsUnique: true,
				},
				{
					Name:     SourceHashAnnotation,
					Value:    string(hash[:]),
					IsUnique: true,
				},
				{
					Name:     SourceFormatAnnotation,
					Value:    sbomReader.Options.Format.Type(),
					IsUnique: true,
				},
			},
		},
	}

	if err := backend.Store(document, opts); err != nil {
		return nil, fmt.Errorf("storing document %s: %w", document.GetMetadata().GetId(), err)
	}

	return document, nil
}

// GetDocumentByID retrieves a protobom Document with the specified ID from the database.
func (backend *Backend) GetDocumentByID(id string) (*sbom.Document, error) {
	document, err := backend.Retrieve(id, nil)
	if err != nil {
		backend.Logger.Warn("Document could not be retrieved", "id", id, "err", err)

		return nil, fmt.Errorf("failed to retrieve document: %w", err)
	}

	return document, nil
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
