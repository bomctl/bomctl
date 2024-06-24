// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/fetch/oci/oci.go
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
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	orasAuth "oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errMultipleSBOMs = errors.New("more than one SBOM document identified in OCI image")

type Fetcher struct{}

func (fetcher *Fetcher) Name() string {
	return "OCI"
}

func (fetcher *Fetcher) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^%s%s%s%s%s$",
			`((?P<scheme>oci|docker)(?:-archive)?:\/\/)?`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`(?P<hostname>[^@\/?#:]+)(?::(?P<port>\d+))?`,
			`(?:\/(?P<path>[^:@]+))`,
			`((?::(?P<tag>[^@]+))|(?:@(?P<digest>sha256:[A-Fa-f0-9]{64})))?`,
		),
	)
}

func (fetcher *Fetcher) Parse(fetchURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := fetcher.RegExp()
	match := pattern.FindStringSubmatch(fetchURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	if results["scheme"] == "docker" || results["scheme"] == "" {
		results["scheme"] = "oci"
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname", "path"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	// One and only one of `tag` or `digest` must be present.
	tag, ok := results["tag"]
	hasTag := ok && tag != ""

	digest, ok := results["digest"]
	hasDigest := ok && digest != ""

	// If both `tag` and `digest` are present, or neither are.
	if hasTag == hasDigest {
		return nil
	}

	return &url.ParsedURL{
		Scheme:   results["scheme"],
		Username: results["username"],
		Password: results["password"],
		Hostname: results["hostname"],
		Port:     results["port"],
		Path:     results["path"],
		Tag:      results["tag"],
		Digest:   results["digest"],
	}
}

func (fetcher *Fetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) ([]byte, error) {
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

	if manifestDescriptor, err = fetchManifestDescriptor(ctx, memStore, repo, parsedURL.Tag); err != nil {
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
		repo.Client = &orasAuth.Client{
			Client: retry.DefaultClient,
			Cache:  orasAuth.DefaultCache,
			Credential: orasAuth.StaticCredential(parsedURL.Hostname, orasAuth.Credential{
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
