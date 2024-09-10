// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/fetch/fetch.go
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
package fetch

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func Fetch(sbomURL string, opts *options.FetchOptions) error {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	doc, err := GetRemoteDocument(sbomURL, opts)
	if err != nil {
		return err
	}

	// Insert fetched document data into database.
	if err := backend.AddDocument(doc); err != nil {
		return fmt.Errorf("error adding document: %w", err)
	}

	if opts.Alias != "" {
		if err := backend.SetAlias(doc.GetMetadata().GetId(), opts.Alias); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if err := backend.AddAnnotations(doc.GetMetadata().GetId(), db.TagAnnotation, opts.Tags...); err != nil {
		return fmt.Errorf("failed to set tags: %w", err)
	}

	// Fetch externally referenced BOMs
	return fetchExternalReferences(doc, backend, opts)
}

func GetRemoteDocument(sbomURL string, opts *options.FetchOptions) (*sbom.Document, error) {
	fetcher, err := client.New(sbomURL)
	if err != nil {
		return nil, fmt.Errorf("creating fetch client: %w", err)
	}

	opts.Logger.Info(fmt.Sprintf("Fetching from %s URL", fetcher.Name()), "url", sbomURL)

	sbomData, err := fetcher.Fetch(sbomURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", sbomURL, err)
	}

	if opts.OutputFile != nil {
		// Write the SBOM document bytes to file.
		if _, err = io.Copy(opts.OutputFile, bytes.NewReader(sbomData)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", opts.OutputFile.Name(), err)
		}
	}

	sbomReader := reader.New()

	document, err := sbomReader.ParseStream(bytes.NewReader(sbomData))
	if err != nil {
		return nil, fmt.Errorf("error parsing SBOM file content: %w", err)
	}

	return document, nil
}

func fetchExternalReferences(document *sbom.Document, backend *db.Backend, opts *options.FetchOptions) error {
	extRefs, err := backend.GetExternalReferencesByDocumentID(document.GetMetadata().GetId(), "BOM")
	if err != nil {
		return fmt.Errorf("error getting external references: %w", err)
	}

	for _, ref := range extRefs {
		extRefsOpt := *opts
		if extRefsOpt.OutputFile != nil {
			out, err := getRefFile(opts.OutputFile)
			if err != nil {
				return err
			}

			extRefsOpt.OutputFile = out
			defer extRefsOpt.OutputFile.Close() //nolint:revive
		}

		if err := Fetch(ref.GetUrl(), &extRefsOpt); err != nil {
			return err
		}
	}

	return nil
}

func getRefFile(parentFile *os.File) (*os.File, error) {
	idx := 0

	// Matches base filename, excluding extension
	baseFilename := regexp.MustCompile(`^([^\.]+)?`).FindString(filepath.Base(parentFile.Name()))

	suffix := regexp.MustCompile(`^.*-(\d+)`).FindString(baseFilename)

	if suffix != "" {
		var err error

		idx, err = strconv.Atoi(suffix)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	idx++

	outputFile := fmt.Sprintf("%s-%d%s",
		filepath.Join(filepath.Dir(parentFile.Name()), baseFilename),
		idx,
		filepath.Ext(parentFile.Name()),
	)

	refOutput, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return refOutput, nil
}
