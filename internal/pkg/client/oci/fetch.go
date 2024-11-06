// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/fetch.go
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
	"fmt"
	"slices"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	url := client.Parse(fetchURL)
	auth := netutil.NewBasicAuth(url.Username, url.Password)

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return nil, fmt.Errorf("failed to set auth: %w", err)
		}
	}

	err := client.createRepository(url, auth, opts.Options)
	if err != nil {
		return nil, err
	}

	ref := url.Tag
	if ref == "" {
		ref = url.Digest
	}

	copyOpts := oras.CopyOptions{CopyGraphOptions: oras.CopyGraphOptions{FindSuccessors: nil}}

	manifest, err := oras.Copy(client.ctx, client.repo, ref, client.store, ref, copyOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest descriptor: %w", err)
	}

	opts.Logger.Debug("Fetched manifest", "descriptor", descriptorJSON(&manifest))

	sbomDescriptor, err := client.getSBOMDescriptor(&manifest)
	if err != nil {
		return nil, err
	}

	opts.Logger.Debug("Found SBOM", "descriptor", descriptorJSON(&sbomDescriptor))

	return client.pullSBOM(&sbomDescriptor)
}

func (client *Client) getSBOMDescriptor(manifest *ocispec.Descriptor) (ocispec.Descriptor, error) {
	// Get all "children" of the manifest
	successors, err := content.Successors(client.ctx, client.store, *manifest)
	if err != nil {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("failed to retrieve manifest layers: %w", err)
	}

	sbomDescriptor := ocispec.DescriptorEmptyJSON
	sbomDigests := []string{}

	for _, descriptor := range successors {
		if slices.ContainsFunc([]string{"application/vnd.cyclonedx", "application/spdx", "text/spdx"},
			func(s string) bool {
				return strings.HasPrefix(descriptor.MediaType, s)
			},
		) {
			sbomDescriptor = descriptor
			sbomDigests = append(sbomDigests, descriptor.Digest.String())
		}
	}

	// Error if more than one SBOM identified
	if len(sbomDigests) > 1 {
		return ocispec.DescriptorEmptyJSON, fmt.Errorf("%w.\n\t%s", ErrMultipleSBOMs, strings.Join(
			append([]string{"Specify one of the following digests in the fetch URL:"}, sbomDigests...),
			"\n\t\t",
		))
	}

	return sbomDescriptor, nil
}

func (client *Client) pullSBOM(sbomDescriptor *ocispec.Descriptor) ([]byte, error) {
	sbomData, err := content.FetchAll(client.ctx, client.store, *sbomDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SBOM data: %w", err)
	}

	return sbomData, nil
}
