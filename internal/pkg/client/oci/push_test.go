// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/push_test.go
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
	"bytes"
	"fmt"
	neturl "net/url"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/protobom/protobom/pkg/formats"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"

	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (ocs *ociClientSuite) TestClient_AddFile() {
	serverURL, err := neturl.Parse(ocs.Server.URL)
	ocs.Require().NoError(err)

	ocs.Require().NoError(
		ocs.Client.PreparePush(
			fmt.Sprintf("%s/%s:%s", serverURL.Host, repoName, manifestTag),
			&options.PushOptions{Options: ocs.Options},
		),
	)

	ocs.Repo().Client = &orasauth.Client{Client: ocs.Server.Client()}

	// Test adding all SBOM files to artifact archive.
	for _, document := range ocs.documents {
		ocs.Require().NoError(ocs.Client.AddFile(
			fmt.Sprintf("%s/%s:%s", serverURL.Host, repoName, manifestTag),
			document.GetMetadata().GetId(),
			&options.PushOptions{Options: ocs.Options, Format: formats.SPDX23JSON},
		))
	}

	ocs.Require().Len(ocs.Descriptors(), len(ocs.documents))

	for _, descriptor := range ocs.Descriptors() {
		exists, err := ocs.Client.Store().Exists(ocs.Context(), descriptor)
		ocs.Require().NoError(err)
		ocs.True(exists)
	}
}

func (ocs *ociClientSuite) TestClient_Push() {
	serverURL, err := neturl.Parse(ocs.Server.URL)
	ocs.Require().NoError(err)

	ocs.Repo().Client = &orasauth.Client{Client: ocs.Server.Client()}
	pushURL := fmt.Sprintf("%s/%s:%s", serverURL.Host, repoName, manifestTag+"-single")
	annotations := oci.Annotations{ocispec.AnnotationCreated: created}

	for idx := range ocs.sbomBlobs {
		_, err := ocs.Client.PushBlob("application/spdx+json", bytes.NewBuffer(ocs.sbomBlobs[idx]), annotations)
		ocs.Require().NoError(err)
	}

	_, _, err = ocs.Client.PackManifest(manifestTag+"-single", annotations)
	ocs.Require().NoError(err)

	ocs.Require().NoError(ocs.Client.Push(pushURL, &options.PushOptions{
		Options: ocs.Options,
		Format:  formats.SPDX23JSON,
		UseTree: false,
	}))
}
