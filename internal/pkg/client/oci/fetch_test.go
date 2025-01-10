// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/fetch_test.go
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

package oci_test

import (
	"fmt"
	neturl "net/url"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (ocs *ociClientSuite) TestClient_Fetch() {
	opts := &options.FetchOptions{Options: ocs.Options}
	descriptors := []ocispec.Descriptor{configDesc}
	serverURL, err := neturl.Parse(ocs.Server.URL)
	ocs.Require().NoError(err)

	for idx := range ocs.sbomBlobs {
		blob := ocs.sbomBlobs[idx]
		blobDigest := digest.FromBytes(blob)
		desc := ocispec.Descriptor{
			MediaType:   "application/spdx+json",
			Digest:      blobDigest,
			Size:        int64(len(blob)),
			Annotations: map[string]string{ocispec.AnnotationCreated: created},
		}

		ocs.ociTestRepository.blobs[string(blobDigest)] = blob
		ocs.ociTestRepository.descriptors[string(blobDigest)] = desc

		descriptors = append(descriptors, desc)
	}

	for _, data := range []struct {
		name        string
		tag         string
		expectedErr error
		descriptors []ocispec.Descriptor
	}{
		{
			name:        "No SBOMs",
			tag:         manifestTag + "-empty",
			expectedErr: nil,
			descriptors: []ocispec.Descriptor{configDesc},
		},
		{
			name:        "Single SBOM",
			tag:         manifestTag + "-single",
			expectedErr: nil,
			descriptors: descriptors[0:2],
		},
		{
			name:        "Multiple SBOMs",
			tag:         manifestTag + "-multiple",
			expectedErr: oci.ErrMultipleSBOMs,
			descriptors: descriptors[0:3],
		},
	} {
		ocs.Run(data.name, func() {
			fetchURL := fmt.Sprintf("oci://%s/%s?ref=%s", serverURL.Host, repoName, data.tag)

			ocs.Require().NoError(ocs.packManifest(data.tag, data.descriptors...))

			if _, err := ocs.Fetch(fetchURL, opts); data.expectedErr != nil {
				ocs.Require().ErrorContains(err, data.expectedErr.Error())
			} else {
				ocs.Require().NoError(err)
			}
		})
	}
}
