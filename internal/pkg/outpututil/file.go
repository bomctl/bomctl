// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/outpututil/file.go
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

package outpututil

import (
	"fmt"
	"io"
	"os"

	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func checkIfModified(document *sbom.Document, backend *db.Backend) (bool, error) {
	sourceContent, err := backend.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.SourceDataAnnotation)
	if err != nil {
		return true, fmt.Errorf("%w", err)
	}

	return (sourceContent == ""), nil
}

func matchesOriginFormat(doc *sbom.Document, format formats.Format) bool {
	return format == formats.Format(doc.GetMetadata().GetSourceData().GetFormat())
}

func writeOriginStream(document *sbom.Document, backend *db.Backend, stream io.WriteCloser) error {
	sourceContent, err := backend.GetDocumentUniqueAnnotation(document.GetMetadata().GetId(), db.SourceDataAnnotation)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if _, err := stream.Write([]byte(sourceContent)); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func WriteStream(document *sbom.Document, format formats.Format, opts *options.Options, stream io.WriteCloser) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	modified, err := checkIfModified(document, backend)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if format == db.OriginalFormat || (!modified && matchesOriginFormat(document, format)) {
		return writeOriginStream(document, backend, stream)
	}

	wrtr := writer.New(writer.WithFormat(format))

	err = wrtr.WriteStream(document, stream)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func WriteFile(document *sbom.Document, format formats.Format, opts *options.Options, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer file.Close()

	return WriteStream(document, format, opts, file)
}
