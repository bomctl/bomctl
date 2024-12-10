// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/fetch/fetch.go
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

package fetch

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

func Fetch(sbomURL string, opts *options.FetchOptions) (*sbom.Document, error) {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

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

	document, err := saveDocument(sbomData, backend, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	if err := backend.SetDocumentUniqueAnnotation(
		document.GetMetadata().GetId(), db.SourceURLAnnotation, sbomURL,
	); err != nil {
		return nil, fmt.Errorf("applying unique annotation %s to %s: %w",
			db.SourceURLAnnotation, document.GetMetadata().GetId(), err,
		)
	}

	// Fetch externally referenced BOMs
	return document, fetchExternalReferences(document, backend, opts)
}

func fetchExternalReferences(document *sbom.Document, backend *db.Backend, opts *options.FetchOptions) error {
	extRefs, err := backend.GetExternalReferencesByDocumentID(document.GetMetadata().GetId(), "BOM")
	if err != nil {
		return fmt.Errorf("error getting external references: %w", err)
	}

	for _, ref := range extRefs {
		extRefOpts := *opts
		if extRefOpts.OutputFile != nil {
			out, err := getRefFile(opts.OutputFile)
			if err != nil {
				return err
			}

			extRefOpts.OutputFile = out
			defer extRefOpts.OutputFile.Close() //revive:disable:defer
		}

		extRefDoc, err := Fetch(ref.GetUrl(), &extRefOpts)
		if err != nil {
			return err
		}

		// Search through all nodes for the corresponding external reference source.
		extRefNode, err := sliceutil.Next(document.GetNodeList().GetNodes(), func(n *sbom.Node) bool {
			return sliceutil.Any(n.GetExternalReferences(), func(er *sbom.ExternalReference) bool {
				return er.GetUrl() == ref.GetUrl()
			})
		})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if err := backend.AddNodeAnnotations(
			extRefNode.GetId(),
			db.LinkToAnnotation,
			extRefDoc.GetMetadata().GetId(),
		); err != nil {
			return fmt.Errorf("%w", err)
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

func saveDocument(data []byte, backend *db.Backend, opts *options.FetchOptions) (*sbom.Document, error) {
	// Insert fetched document data into database.
	document, err := backend.AddDocument(data, db.WithSourceDocumentAnnotations(data))
	if err != nil {
		return nil, fmt.Errorf("adding document: %w", err)
	}

	if opts.Alias != "" {
		if err := backend.SetAlias(document.GetMetadata().GetId(), opts.Alias, false); err != nil {
			opts.Logger.Warn("Alias could not be set.", "err", err)
		}
	}

	if err := backend.AddDocumentAnnotations(
		document.GetMetadata().GetId(), db.TagAnnotation, opts.Tags...,
	); err != nil {
		opts.Logger.Warn("Tag(s) could not be set.", "err", err)
	}

	// Fetch externally referenced BOMs
	return document, nil
}
