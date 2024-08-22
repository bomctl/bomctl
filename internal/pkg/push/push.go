// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/push/push.go
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

func Push(sbomID, destPath string, opts *options.PushOptions) error {
	opts.Logger.Info("Pushing Document", "sbomID", sbomID)

	// create appropriate push client based on user provided destination
	pClient, err := client.New(destPath)
	if err != nil {
		return fmt.Errorf("creating push client: %w", err)
	}

	opts.Logger.Info(fmt.Sprintf("Pushing to %s URL", pClient.Name()), "url", destPath)

	// push sbomid to destpath via client
	err = pClient.Push(sbomID, destPath, opts)
	if err != nil {
		return fmt.Errorf("failed to push to %s: %w", destPath, err)
	}

	// If user wants to recurse the sbom tree and push all, do so
	if opts.UseTree {
		err := getExtRefDocs(sbomID, destPath, opts)
		if err != nil {
			return fmt.Errorf("failed to fetch external ref boms for %s: %w", sbomID, err)
		}
	}

	return nil
}

func getExtRefDocs(sbomID, destPath string, opts *options.PushOptions) error {
	opts.Logger.Info("Fetching External Ref Boms for Document", "sbomID", sbomID)

	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	extRefs, err := backend.GetExternalReferencesByDocumentID(sbomID, "BOM")
	if err != nil {
		return fmt.Errorf("error getting external references: %w", err)
	}

	// for each external ref fetch and push external ref bom document
	for _, ref := range extRefs {
		err := pushExtRefDoc(ref, backend, destPath, opts)
		if err != nil {
			return fmt.Errorf("error pulling external ref file: %w", err)
		}
	}

	return nil
}

func checkForExistingDoc(be *db.Backend, doc *sbom.Document) (id, name string, err error) {
	// checks local db for fetched document identifiers
	// returns local data if found, otherwise uses fetched data
	existingDoc, err := be.GetDocumentByID(doc.Metadata.Id)
	if err != nil {
		if err = be.AddDocument(doc); err != nil {
			return "", "", fmt.Errorf("failed to store document: %w", err)
		}

		return doc.Metadata.Id, doc.Metadata.Name, nil
	}

	return existingDoc.Metadata.Id, existingDoc.Metadata.Name, nil
}

func getExtRefPath(destPath, docID, docName string) string {
	// generate destination path to push to
	// based on what we know about the bom and the requested dest url
	// will push to same path (dir) as the origin pushed bom
	// but with name or id from fecth doc and requested format ext
	ext := filepath.Ext(destPath)
	base := filepath.Base(destPath)
	destDir := destPath[:len(destPath)-len(base)]

	// document name doesnt exist, use id
	if docName == "" {
		return (destDir + docID + ext)
	}

	fileName := strings.ReplaceAll(docName, ".", "_")

	return (destDir + fileName + ext)
}

func pushExtRefDoc(ref *sbom.ExternalReference, be *db.Backend, destPath string, opts *options.PushOptions) error {
	opts.Logger.Info("Fetching External Ref Bom from URL", "refUrl", ref.Url)

	// Parse push options into fetch
	fetchOpts := &options.FetchOptions{
		UseNetRC: opts.UseNetRC,
		Options:  opts.Options,
	}

	// call fetch wrapper function to fetch extref doc object
	doc, err := fetch.GetRemoteDocument(ref.Url, fetchOpts)
	if err != nil {
		return fmt.Errorf("error fetching external reference docs: %w", err)
	}

	// check local db to see if this file already exists
	docID, docName, err := checkForExistingDoc(be, doc)
	if err != nil {
		return fmt.Errorf("failed to import external ref bom: %w", err)
	}

	// generate deestination path for ext ref doc based on provided dest pah
	extRefDestPath := getExtRefPath(destPath, docID, docName)
	opts.Logger.Info("pushing External Ref Bom Document", "destPath", extRefDestPath)

	// push extref bom, calling top level push cmd
	// Which will check the pushed ext ref bom for ext ref boms
	// And recursively fetch and push the entire sbom tree
	err = Push(docID, extRefDestPath, opts)
	if err != nil {
		return fmt.Errorf("failed to push to %s: %w", extRefDestPath, err)
	}

	return nil
}
