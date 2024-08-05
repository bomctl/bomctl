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
	"path/filepath"

	"github.com/protobom/protobom/pkg/reader"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

type ImportOptions struct {
	*options.Options
	InputFiles []*os.File
}

func Import(opts *ImportOptions) error {
	backend := db.NewBackend().
		Debug(opts.Debug).
		WithDatabaseFile(filepath.Join(opts.CacheDir, db.DatabaseFile)).
		WithLogger(opts.Logger)

	if err := backend.InitClient(); err != nil {
		return fmt.Errorf("failed to initialize backend client: %w", err)
	}

	defer backend.CloseClient()

	sbomReader := reader.New()

	for idx := range opts.InputFiles {
		data, err := io.ReadAll(opts.InputFiles[idx])
		if err != nil {
			return fmt.Errorf("failed to read from %s: %w", opts.InputFiles[idx].Name(), err)
		}

		document, err := sbomReader.ParseStream(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", opts.InputFiles[idx].Name(), err)
		}

		if err := backend.AddDocument(document); err != nil {
			return fmt.Errorf("failed to store document: %w", err)
		}
	}

	return nil
}
