// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/push/push.go
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

package push

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/fetch"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func NewPusher(url string) (client.Pusher, error) {
	c, err := client.New(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", client.ErrUnsupportedURL, url)
	}

	pusher, ok := c.(client.Pusher)
	if !ok {
		return nil, fmt.Errorf("%w: %s", client.ErrUnsupportedURL, url)
	}

	return pusher, nil
}

func Push(sbomID, pushURL string, opts *options.PushOptions) error {
	opts.Logger.Info("Pushing document", "id", sbomID)

	// Create appropriate push client based on user provided destination.
	pushClient, err := NewPusher(pushURL)
	if err != nil {
		return fmt.Errorf("creating push client: %w", err)
	}

	opts.Logger.Info(fmt.Sprintf("Pushing to %s URL", pushClient.Name()), "url", pushURL)

	if err := pushClient.PreparePush(pushURL, opts); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := pushClient.AddFile(pushURL, sbomID, opts); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Recurse the SBOM tree and push all.
	if opts.UseTree {
		if err := addExternalReferenceFiles(sbomID, pushURL, pushClient, opts); err != nil {
			return err
		}
	}

	// Finalize and push.
	if err := pushClient.Push(pushURL, opts); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func addExternalReferenceFiles(sbomID, pushURL string, pushClient client.Pusher, opts *options.PushOptions) error {
	extRefs, err := getExternalReferences(sbomID, opts)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	extRefDocs, err := resolveExternalReferences(extRefs, opts)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	for _, document := range extRefDocs {
		extRefURL := getExtRefPath(pushURL, document.GetMetadata().GetId(), document.GetMetadata().GetName(), opts)

		if err := pushClient.AddFile(extRefURL, document.GetMetadata().GetId(), opts); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func getExternalReferences(sbomID string, opts *options.PushOptions) ([]*sbom.ExternalReference, error) {
	opts.Logger.Info("Fetching external reference SBOMs", "id", sbomID)

	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	extRefs, err := backend.GetExternalReferencesByDocumentID(sbomID, "BOM")
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return extRefs, nil
}

func resolveExternalReferences(refs []*sbom.ExternalReference, opts *options.PushOptions) ([]*sbom.Document, error) {
	documents := []*sbom.Document{}

	fetchOpts := &options.FetchOptions{
		UseNetRC: opts.UseNetRC,
		Options:  opts.Options,
	}

	for _, ref := range refs {
		document, err := fetch.Fetch(ref.GetUrl(), fetchOpts)
		if err != nil {
			return nil, fmt.Errorf("fetching external reference document: %w", err)
		}

		documents = append(documents, document)

		extRefs, err := getExternalReferences(document.GetMetadata().GetId(), opts)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		extRefDocs, err := resolveExternalReferences(extRefs, opts)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		documents = append(documents, extRefDocs...)
	}

	return documents, nil
}

// generate destination path to push to based on what
// we know about the bom and the requested dest url
// pushes to same path (dir) as the origin pushed bom
// but with name or id from fetch doc and requested format ext.
func getExtRefPath(destPath, docID, docName string, opts *options.PushOptions) string {
	ext := filepath.Ext(destPath)
	destDir := filepath.Dir(destPath)

	// Document name doesn't exist, use ID.
	if docName == "" {
		return filepath.Join(destDir, fmt.Sprintf("%s%s", docID, ext))
	}

	fileName := fmt.Sprintf("%s%s", strings.ReplaceAll(docName, ".", "_"), ext)

	opts.Logger.Info("External reference SBOM", "name", fileName)

	return filepath.Join(destDir, fileName)
}
