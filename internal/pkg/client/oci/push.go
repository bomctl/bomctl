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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/writer"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const schemaVersion = 2

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

	// Add annotation to save file name.
	annotations := map[string]string{}
	if url := client.Parse(pushURL); url != nil {
		annotations[ocispec.AnnotationTitle] = path.Base(url.Path)
	}

	sbomDescriptor, err := client.pushBlob(getMediaType(opts), buf.Buffer, annotations)
	if err != nil {
		return fmt.Errorf("pushing to memory store: %w", err)
	}

	opts.Logger.Debug("Pushed artifact", "digest", sbomDescriptor.Digest)

	client.descriptors = append(client.descriptors, sbomDescriptor)

	return nil
}

func (client *Client) PreparePush(pushURL string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	if url == nil {
		return fmt.Errorf("%w", netutil.ErrParsingURL)
	}

	auth := &netutil.BasicAuth{Username: url.Username, Password: url.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return fmt.Errorf("setting .netrc auth: %w", err)
		}
	}

	return client.createRepository(url, auth, opts.Options)
}

func (client *Client) Push(pushURL string, opts *options.PushOptions) error {
	defer func() {
		clear(client.descriptors)
		client.repo = nil
		client.store = nil
	}()

	tag := ""
	if url := client.Parse(pushURL); url != nil {
		tag = url.Tag
		if tag == "" {
			tag = "latest"
		}
	}

	manifestDesc, manifestBytes, err := client.generateManifest(nil)
	if err != nil {
		return err
	}

	if err := client.store.Push(client.ctx, manifestDesc, bytes.NewReader(manifestBytes)); err != nil {
		return fmt.Errorf("pushing manifest to memory store: %w", err)
	}

	if err := client.store.Tag(client.ctx, manifestDesc, tag); err != nil {
		return fmt.Errorf("tagging manifest: %w", err)
	}

	opts.Logger.Debug("Packed manifest", "descriptor", descriptorJSON(&manifestDesc), "data", string(manifestBytes))

	if _, err := oras.Copy(client.ctx, client.store, tag, client.repo, tag, oras.DefaultCopyOptions); err != nil {
		return fmt.Errorf("pushing %s with digest %s to remote repository: %w", tag, string(manifestDesc.Digest), err)
	}

	opts.Logger.Debug("Copied manifest", "url", pushURL)

	return nil
}

func (client *Client) generateManifest(annotations map[string]string) (ocispec.Descriptor, []byte, error) {
	configDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    ocispec.DescriptorEmptyJSON.Digest,
		Size:      ocispec.DescriptorEmptyJSON.Size,
	}

	client.descriptors = append([]ocispec.Descriptor{ocispec.DescriptorEmptyJSON}, client.descriptors...)

	if err := client.store.Push(
		client.ctx,
		configDesc,
		bytes.NewReader(ocispec.DescriptorEmptyJSON.Data),
	); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return ocispec.Descriptor{}, []byte{}, fmt.Errorf("pushing config blob: %w", err)
	}

	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Add the "org.opencontainers.image.created" annotation to the blob descriptor if not provided.
	if _, ok := annotations[ocispec.AnnotationCreated]; !ok {
		annotations[ocispec.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)
	}

	manifest := ocispec.Manifest{
		Versioned:   specs.Versioned{SchemaVersion: schemaVersion},
		Config:      configDesc,
		Layers:      client.descriptors,
		Annotations: annotations,
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return ocispec.Descriptor{}, manifestBytes, fmt.Errorf("marshaling manifest: %w", err)
	}

	manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, manifestBytes)

	return manifestDesc, manifestBytes, nil
}

func (client *Client) pushBlob(
	mediaType string, data io.Reader, annotations map[string]string,
) (ocispec.Descriptor, error) {
	blob := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(blob, data); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("reading blob content: %w", err)
	}

	// Add the "org.opencontainers.image.created" annotation to the blob descriptor if not provided.
	if _, ok := annotations[ocispec.AnnotationCreated]; !ok {
		annotations[ocispec.AnnotationCreated] = time.Now().UTC().Format(time.RFC3339)
	}

	desc := ocispec.Descriptor{
		MediaType:   mediaType,
		Digest:      digest.FromBytes(blob.Bytes()),
		Size:        int64(blob.Len()),
		Annotations: annotations,
	}

	// Push SBOM descriptor blob to target.
	if err := client.store.Push(client.ctx, desc, blob); err != nil && !errors.Is(err, errdef.ErrAlreadyExists) {
		return ocispec.Descriptor{}, fmt.Errorf("pushing blob: %w", err)
	}

	return desc, nil
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
