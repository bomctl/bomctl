// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
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

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/utils"
	"github.com/bomctl/bomctl/internal/pkg/utils/format"
	"github.com/charmbracelet/log"

	"github.com/protobom/protobom/pkg/writer"
)

type (
	ExportOptions struct {
		Logger       *log.Logger
		OutputFile   *os.File
		FormatString string
		Encoding     string
		CacheDir     string
		ConfigFile   string
	}
)

func Export(sbomID string, opts *ExportOptions) error {
	// func Exec(sbomID, outputFile, fs, encoding string) error {
	logger := utils.NewLogger("export")

	logger.Info(fmt.Sprintf("Saving %s SBOM ID", sbomID))
	backend := db.NewBackend()

	document, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	parsedFormat, err := format.Parse(opts.FormatString, opts.Encoding)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	writer := writer.New(
		writer.WithFormat(parsedFormat),
	)

	if err := writer.WriteFile(document, opts.OutputFile.Name()); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
