// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/import/import.go
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

package imprt

import (
	"fmt"
	"io"
	"os"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func Import(opts *options.ImportOptions) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for idx := range opts.InputFiles {
		alias := ""
		if idx < len(opts.Alias) {
			alias = opts.Alias[idx]
		}

		if err := saveDocument(backend, opts.InputFiles[idx], alias, opts); err != nil {
			return fmt.Errorf("importing document: %w", err)
		}
	}

	return nil
}

func saveDocument(backend *db.Backend, documentFile *os.File, alias string, opts *options.ImportOptions) error {
	data, err := io.ReadAll(documentFile)
	if err != nil {
		return fmt.Errorf("failed to read from %s: %w", documentFile.Name(), err)
	}

	document, err := backend.AddDocument(data)
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	if alias != "" {
		if err := backend.SetAlias(document.GetMetadata().GetId(), alias, false); err != nil {
			opts.Logger.Warn("Alias could not be set.", "err", err)
		}
	}

	if err := backend.AddDocumentAnnotations(document.GetMetadata().GetId(), db.TagAnnotation, opts.Tags...); err != nil {
		opts.Logger.Warn("Tag(s) could not be set.", "err", err)
	}

	return nil
}
