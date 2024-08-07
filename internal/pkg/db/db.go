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
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"

	"github.com/bomctl/bomctl/internal/pkg/utils"
)

const DatabaseFile string = "bomctl.db"

type (
	Backend struct {
		*ent.Backend
		Logger *log.Logger
	}

	Option func(*Backend)
)

func NewBackend(opts ...Option) *Backend {
	backend := &Backend{
		Backend: ent.NewBackend(),
		Logger:  utils.NewLogger("db"),
	}

	for _, opt := range opts {
		opt(backend)
	}

	return backend
}

// AddDocument adds the protobom Document to the database.
func (backend *Backend) AddDocument(document *sbom.Document) error {
	if err := backend.Store(document, nil); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
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

// Debug enables debug logging for all database transactions.
func (backend *Backend) Debug(debug bool) *Backend {
	backend.Options.Debug = debug

	return backend
}

// WithLogger sets the logger for the backend.
func (backend *Backend) WithLogger(logger *log.Logger) *Backend {
	backend.Logger = logger

	return backend
}

// WithDatabaseFile sets the database file for the backend.
func (backend *Backend) WithDatabaseFile(file string) *Backend {
	backend.Options.DatabaseFile = file

	return backend
}

// Debug enables debug logging for all database transactions.
func Debug(debug bool) Option {
	return func(backend *Backend) {
		backend.Debug(debug)
	}
}

// WithLogger sets the logger for the backend.
func WithLogger(logger *log.Logger) Option {
	return func(backend *Backend) {
		backend.WithLogger(logger)
	}
}

// WithDatabaseFile sets the database file for the backend.
func WithDatabaseFile(file string) Option {
	return func(backend *Backend) {
		backend.WithDatabaseFile(file)
	}
}
