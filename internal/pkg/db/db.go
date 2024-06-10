// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
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
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/storage/backends/ent"
	"github.com/spf13/viper"
)

const DatabaseFile string = "bomctl.db"

var (
	backend *ent.Backend
	logger  *log.Logger
)

// AddDocument adds the protobom Document to the database.
func AddDocument(document *sbom.Document) error {
	if backend == nil {
		if err := initBackend(); err != nil {
			return fmt.Errorf("initBackend: %w", err)
		}
	}

	if err := backend.Store(document, nil); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	return nil
}

// GetDocumentByID retrieves a protobom Document with the specified ID from the database.
func GetDocumentByID(id string) (*sbom.Document, error) {
	if backend == nil {
		if err := initBackend(); err != nil {
			return nil, fmt.Errorf("initBackend: %w", err)
		}
	}

	document, err := backend.Retrieve(id, nil)
	if err != nil {
		logger.Warn("Document could not be retrieved", "id", id, "err", err)

		return nil, fmt.Errorf("failed to retrieve document: %w", err)
	}

	return document, nil
}

// GetExternalReferencesByID returns all ExternalReferences of type "BOM" in an SBOM document.
func GetExternalReferencesByID(id string) (refs []*sbom.ExternalReference) {
	refs = []*sbom.ExternalReference{}

	document, err := GetDocumentByID(id)
	if err != nil {
		return
	}

	for _, node := range document.GetNodeList().GetNodes() {
		for _, ref := range node.GetExternalReferences() {
			if ref.Type == sbom.ExternalReference_BOM {
				refs = append(refs, ref)
			}
		}
	}

	return
}

func initBackend() error {
	cacheDir := viper.GetString("cache_dir")
	backend = ent.NewBackend().WithDatabaseFile(filepath.Join(cacheDir, DatabaseFile))

	if err := backend.InitClient(); err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	return nil
}
