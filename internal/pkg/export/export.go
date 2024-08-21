// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/export/export.go
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
package export

import (
	"fmt"
	"os"

	"github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func Export(sbomID string, opts *options.ExportOptions) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	backend.Logger.Info("Exporting document", "sbomID", sbomID)

	wr := writer.New(writer.WithFormat(opts.Format))

	document, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if opts.OutputFile != nil {
		// Write the SBOM document bytes to file.
		if err := wr.WriteFile(document, opts.OutputFile.Name()); err != nil {
			return fmt.Errorf("%w", err)
		}
	} else {
		// Write the SBOM document bytes to stdout.
		if err := wr.WriteStream(document, os.Stdout); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
