// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/fetch.go
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
package oci

import (
	"context"
	"fmt"
	"slices"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

func (client *Client) Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	parsedURL := client.Parse(fetchURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(parsedURL.Hostname); err != nil {
			return nil, fmt.Errorf("failed to set auth: %w", err)
		}
	}

	var (
		err                                error
		manifestDescriptor, sbomDescriptor *ocispec.Descriptor
		repo                               *remote.Repository
		sbomData                           []byte
		successors                         []ocispec.Descriptor
	)

	if repo, err = createRepository(parsedURL, auth); err != nil {
		return nil, err
	}

	ctx := context.Background()
	memStore := memory.New()

	ref := parsedURL.Tag
	if ref == "" {
		ref = parsedURL.Digest
	}

	if manifestDescriptor, err = fetchManifestDescriptor(ctx, memStore, repo, ref); err != nil {
		return nil, err
	}

	if successors, err = getManifestChildren(ctx, memStore, manifestDescriptor); err != nil {
		return nil, err
	}

	if sbomDescriptor, err = getSBOMDescriptor(successors); err != nil {
		return nil, err
	}

	if sbomData, err = pullSBOM(ctx, memStore, sbomDescriptor); err != nil {
		return nil, err
	}

	return sbomData, nil
}

func createRepository(parsedURL *url.ParsedURL, auth *url.BasicAuth) (*remote.Repository, error) {
	repoPath := strings.Trim(parsedURL.Hostname, "/") + "/" + strings.Trim(parsedURL.Path, "/")

	repo, err := remote.NewRepository(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error creating OCI registry repository %s: %w", repoPath, err)
	}

	if auth != nil {
		repo.Client = &orasauth.Client{
			Client: retry.DefaultClient,
			Cache:  orasauth.DefaultCache,
			Credential: orasauth.StaticCredential(parsedURL.Hostname, orasauth.Credential{
				Username: auth.Username,
				Password: auth.Password,
			}),
		}
	}

	return repo, nil
}

func fetchManifestDescriptor(
	ctx context.Context, memStore *memory.Store, repo *remote.Repository, tag string,
) (*ocispec.Descriptor, error) {
	manifestDescriptor, err := oras.Copy(ctx, repo, tag, memStore, tag,
		oras.CopyOptions{CopyGraphOptions: oras.CopyGraphOptions{FindSuccessors: nil}},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest descriptor: %w", err)
	}

	return &manifestDescriptor, nil
}

func getManifestChildren(
	ctx context.Context, memStore *memory.Store, manifestDescriptor *ocispec.Descriptor,
) ([]ocispec.Descriptor, error) {
	// Get all "children" of the manifest
	successors, err := content.Successors(ctx, memStore, *manifestDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve manifest layers: %w", err)
	}

	return successors, nil
}

func getSBOMDescriptor(successors []ocispec.Descriptor) (*ocispec.Descriptor, error) {
	var (
		sbomDescriptor ocispec.Descriptor
		sbomDigests    []string
	)

	for _, s := range successors {
		if slices.Contains([]string{
			"application/vnd.cyclonedx",
			"application/vnd.cyclonedx+json",
			"application/spdx",
			"application/spdx+json",
			"text/spdx",
		}, s.MediaType) {
			sbomDescriptor = s
			sbomDigests = append(sbomDigests, s.Digest.String())
		}
	}

	// Error if more than one SBOM identified
	if len(sbomDigests) > 1 {
		digestString := strings.Join(
			append([]string{"Specify one of the following digests in the fetch URL:"}, sbomDigests...),
			"\n\t\t",
		)

		return nil, fmt.Errorf("%w.\n\t%s", errMultipleSBOMs, digestString)
	}

	return &sbomDescriptor, nil
}

func pullSBOM(ctx context.Context, memStore *memory.Store, sbomDescriptor *ocispec.Descriptor) ([]byte, error) {
	sbomData, err := content.FetchAll(ctx, memStore, *sbomDescriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SBOM data: %w", err)
	}

	return sbomData, nil
}
