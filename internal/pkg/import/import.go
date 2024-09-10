// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/import/import.go
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
package imprt

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func parseDocument(sbomReader *reader.Reader, inputFile *os.File) (*sbom.Document, error) {
	data, err := io.ReadAll(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read from %s: %w", inputFile.Name(), err)
	}

	sbomDocument, err := sbomReader.ParseStream(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", inputFile.Name(), err)
	}

	return sbomDocument, nil
}

func saveDocument(backend *db.Backend, document *sbom.Document, alias string, tags ...string) error {
	if err := backend.AddDocument(document); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	if alias != "" {
		if err := backend.SetAlias(document.GetMetadata().GetId(), alias); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if err := backend.AddAnnotations(document.GetMetadata().GetId(), db.TagAnnotation, tags...); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	return nil
}

func Import(opts *options.ImportOptions) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	sbomReader := reader.New()

	for idx := range opts.InputFiles {
		document, err := parseDocument(sbomReader, opts.InputFiles[idx])
		if err != nil {
			return fmt.Errorf("failed to read SBOM document %w", err)
		}

		alias := ""
		if idx < len(opts.Alias) {
			alias = opts.Alias[idx]
		}

		if err := saveDocument(backend, document, alias, opts.Tags...); err != nil {
			return fmt.Errorf("failed to save document: %w", err)
		}
	}

	return nil
}
