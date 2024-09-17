// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/push.go
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

package oci

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"path"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/writer"
	oras "oras.land/oras-go/v2"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

type ociClientWriter struct {
	*bytes.Buffer
	*io.PipeReader
}

func (client *Client) AddFile(pushURL, id string, opts *options.PushOptions) error {
	document, err := getDocument(id, opts.Options)
	if err != nil {
		return err
	}

	buf := &ociClientWriter{bytes.NewBuffer([]byte{}), &io.PipeReader{}}

	wr := writer.New(writer.WithFormat(opts.Format))
	if err := wr.WriteStream(document, buf); err != nil {
		return fmt.Errorf("%w", err)
	}

	checksum := sha256.Sum256(buf.Bytes())
	sbomDescriptor := ocispec.Descriptor{
		MediaType: getMediaType(opts),
		Digest:    digest.NewDigestFromBytes(digest.SHA256, checksum[:]),
		Size:      int64(buf.Len()),
	}

	// Add annotation to save file name.
	if parsedURL := client.Parse(pushURL); parsedURL != nil {
		sbomDescriptor.Annotations = map[string]string{"org.opencontainers.image.title": path.Base(parsedURL.Path)}
	}

	// Push SBOM descriptor blob to memory store.
	if err := client.store.Push(client.ctx, sbomDescriptor, buf.Buffer); err != nil {
		return fmt.Errorf("pushing to memory store: %w", err)
	}

	client.descriptors = append(client.descriptors, sbomDescriptor)

	return nil
}

func (client *Client) PreparePush(pushURL string, opts *options.PushOptions) error {
	parsedURL := client.Parse(pushURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(parsedURL.Hostname); err != nil {
			return fmt.Errorf("setting .netrc auth: %w", err)
		}
	}

	return client.createRepository(parsedURL, auth)
}

func (client *Client) Push(pushURL string, opts *options.PushOptions) error {
	defer func() {
		clear(client.descriptors)
		client.repo = nil
		client.store = nil
	}()

	manifest, err := oras.PackManifest(
		client.ctx,
		client.store,
		oras.PackManifestVersion1_1,
		ocispec.MediaTypeImageManifest,
		oras.PackManifestOptions{Layers: client.descriptors},
	)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	tag := ""
	if parsedURL := client.Parse(pushURL); parsedURL != nil {
		tag = parsedURL.Tag
		if tag == "" {
			tag = "latest"
		}
	}

	opts.Logger.Debug("Applying tag", "tag", tag, "digest", manifest.Digest)

	if err := client.store.Tag(client.ctx, manifest, tag); err != nil {
		return fmt.Errorf("%w", err)
	}

	opts.Logger.Debug("Packed manifest", "descriptor", descriptorJSON(&manifest))

	if _, err := oras.Copy(client.ctx, client.store, tag, client.repo, tag, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("%w", err)
	}

	opts.Logger.Debug("Copied manifest", "url", pushURL)

	return nil
}

func getDocument(id string, opts *options.Options) (*sbom.Document, error) {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Retrieve document from database.
	doc, err := backend.GetDocumentByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return doc, nil
}

func getMediaType(opts *options.PushOptions) string {
	opts.Logger.Debug("Getting mediaType for descriptor", "format", opts.Format)

	// Only SPDX JSON encoding is currently supported by protobom, and the media type registered with the
	// IANA has no version parameter (https://www.iana.org/assignments/media-types/application/spdx+json).
	if opts.Format.Type() == formats.SPDXFORMAT {
		return "application/spdx+json"
	}

	return string(opts.Format)
}
