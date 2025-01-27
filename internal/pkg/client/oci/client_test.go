// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/client_test.go
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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/stretchr/testify/suite"
	"oras.land/oras-go/v2/content"
	orasauth "oras.land/oras-go/v2/registry/remote/auth"

	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/testutil"
)

const (
	manifestTag = "v1"
	repoName    = "oci-client-test"
	testSHA     = "sha256:abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF0123456789"
)

var (
	configDesc = ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    ocispec.DescriptorEmptyJSON.Digest,
		Size:      ocispec.DescriptorEmptyJSON.Size,
	}

	created = time.Now().UTC().Format(time.RFC3339)
)

type (
	ociClientSuite struct {
		suite.Suite
		*db.Backend
		*oci.Client
		*ociTestRepository
		*options.Options
		documents    []*sbom.Document
		documentInfo []testutil.DocumentInfo
		sbomBlobs    [][]byte
	}

	ociTestRepository struct {
		*httptest.Server
		ctx         context.Context
		manifests   map[string][]byte
		blobs       map[string][]byte
		descriptors map[string]ocispec.Descriptor
	}
)

func (ocs *ociClientSuite) BeforeTest(_suiteName, _testName string) {
	ocs.Client = &oci.Client{}
	ocs.ociTestRepository = &ociTestRepository{}

	var err error

	ocs.Backend, err = testutil.NewTestBackend()
	ocs.Require().NoError(err, "failed database backend creation")

	ocs.documentInfo, err = testutil.AddTestDocuments(ocs.Backend)
	ocs.Require().NoError(err, "failed database backend setup")

	ocs.Options = options.New().WithContext(context.WithValue(context.Background(), db.BackendKey{}, ocs.Backend))

	ocs.setupOCIRepository()

	serverURL, err := neturl.Parse(ocs.Server.URL)
	ocs.Require().NoError(err)

	url := &netutil.URL{
		Hostname: serverURL.Hostname(),
		Port:     serverURL.Port(),
		Path:     repoName,
		Tag:      manifestTag,
	}

	ocs.Require().NoError(ocs.Client.CreateRepository(url, nil, ocs.Options))

	ocs.ctx = ocs.Context()
	testdataDir := testutil.GetTestdataDir()

	sboms, err := os.ReadDir(testdataDir)
	ocs.Require().NoError(err)

	ocs.ociTestRepository.blobs[string(configDesc.Digest)] = ocispec.DescriptorEmptyJSON.Data
	ocs.ociTestRepository.descriptors[string(configDesc.Digest)] = configDesc

	descriptors := []ocispec.Descriptor{configDesc}

	for idx := range sboms {
		sbomData, err := os.ReadFile(filepath.Join(testdataDir, sboms[idx].Name()))
		if err != nil {
			ocs.T().Fatalf("%v", err)
		}

		doc, err := ocs.Backend.AddDocument(sbomData, db.WithSourceDocumentAnnotations(sbomData))
		if err != nil {
			ocs.FailNow("failed storing document", "err", err)
		}

		desc := ocispec.Descriptor{
			MediaType: "application/spdx+json",
			Digest:    digest.FromBytes(sbomData),
			Size:      int64(len(sbomData)),
		}

		ocs.sbomBlobs = append(ocs.sbomBlobs, sbomData)
		ocs.documents = append(ocs.documents, doc)
		descriptors = append(descriptors, desc)

		ocs.ociTestRepository.blobs[string(desc.Digest)] = sbomData
		ocs.ociTestRepository.descriptors[string(desc.Digest)] = desc
	}

	// Generate manifests for testing
	ocs.Require().NoError(ocs.ociTestRepository.packManifest("v1-empty"))
	ocs.Require().NoError(ocs.ociTestRepository.packManifest("v1-single", descriptors[0]))
	ocs.Require().NoError(ocs.ociTestRepository.packManifest("v1-multiple", descriptors[0:2]...))

	ocs.Repo().Client = &orasauth.Client{Client: ocs.Server.Client()}
}

func (ocs *ociClientSuite) AfterTest(_suiteName, _testName string) {
	ocs.Server.Close()
	ocs.Backend.CloseClient()
}

func (otr *ociTestRepository) blobsHandler() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet || req.Method == http.MethodHead:
			if blob, ok := otr.blobs[req.URL.Path]; ok {
				resp.Header().Set("Content-Length", strconv.Itoa(len(blob)))
				resp.Header().Set("Content-Type", "application/octet-stream")
				resp.Write(blob) //nolint:errcheck

				return
			}

			http.NotFound(resp, req)
		default:
			resp.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (otr *ociTestRepository) blobsUploadsHandler() http.Handler {
	uploadUUID := uuid.NewString()

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPost:
			resp.Header().Set("Content-Length", "0")
			resp.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
			resp.Header().Set("Docker-Upload-UUID", uploadUUID)
			resp.Header().Set("Location", req.RequestURI+uploadUUID)
			resp.Header().Set("Range", "0-0")
			resp.WriteHeader(http.StatusAccepted)
		case req.Method == http.MethodPut && req.URL.Path == uploadUUID:
			digestString, err := neturl.QueryUnescape(req.URL.Query().Get("digest"))
			if err != nil {
				http.Error(resp, req.RequestURI, http.StatusBadRequest)

				return
			}

			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(resp, req.RequestURI, http.StatusBadRequest)

				return
			}

			blobDesc := content.NewDescriptorFromBytes("application/spdx+json", body)

			otr.blobs[digestString] = body
			otr.descriptors[digestString] = blobDesc

			resp.Header().Set("Content-Length", "0")
			resp.Header().Set("Docker-Content-Digest", string(blobDesc.Digest))
			resp.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", repoName, string(blobDesc.Digest)))
			resp.WriteHeader(http.StatusCreated)
		default:
			resp.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (otr *ociTestRepository) manifestsHandler() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodGet:
			body, ok := otr.manifests[req.URL.Path]
			if !ok {
				http.NotFound(resp, req)

				return
			}

			manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, body)

			resp.Header().Set("Content-Length", strconv.Itoa(int(manifestDesc.Size)))
			resp.Header().Set("Content-Type", manifestDesc.MediaType)
			resp.Header().Set("Docker-Content-Digest", string(manifestDesc.Digest))

			resp.Write(body) //nolint:errcheck
		case req.Method == http.MethodHead:
			body, ok := otr.manifests[req.URL.Path]
			if !ok {
				http.NotFound(resp, req)

				return
			}

			manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, body)

			resp.Header().Set("Content-Length", strconv.Itoa(int(manifestDesc.Size)))
			resp.Header().Set("Content-Type", manifestDesc.MediaType)
			resp.Header().Set("Docker-Content-Digest", string(manifestDesc.Digest))
		case req.Method == http.MethodPut:
			manifestBytes, err := io.ReadAll(req.Body)
			if err != nil {
				http.NotFound(resp, req)

				return
			}

			manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, manifestBytes)

			otr.manifests[string(manifestDesc.Digest)] = manifestBytes

			resp.Header().Set("Content-Length", "0")
			resp.Header().Set("Docker-Content-Digest", string(manifestDesc.Digest))
			resp.Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s", repoName, string(manifestDesc.Digest)))
			resp.WriteHeader(http.StatusCreated)
		default:
			resp.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (otr *ociTestRepository) packManifest(tag string, layers ...ocispec.Descriptor) error {
	layers = append([]ocispec.Descriptor{ocispec.DescriptorEmptyJSON}, layers...)

	manifest := ocispec.Manifest{
		Versioned:   specs.Versioned{SchemaVersion: 2},
		Config:      configDesc,
		Layers:      layers,
		Annotations: map[string]string{ocispec.AnnotationCreated: created},
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	manifestDesc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, manifestBytes)

	otr.manifests[tag] = manifestBytes
	otr.manifests[string(manifestDesc.Digest)] = manifestBytes

	return nil
}

func (otr *ociTestRepository) setupOCIRepository() {
	otr.manifests = make(map[string][]byte)
	otr.blobs = make(map[string][]byte)
	otr.descriptors = make(map[string]ocispec.Descriptor)

	ociServeMux := http.NewServeMux()
	otr.Server = httptest.NewTLSServer(ociServeMux)

	prefix := "/v2/oci-client-test"
	ociServeMux.Handle(prefix+"/blobs/", http.StripPrefix(prefix+"/blobs/", otr.blobsHandler()))
	ociServeMux.Handle(prefix+"/blobs/uploads/", http.StripPrefix(prefix+"/blobs/uploads/", otr.blobsUploadsHandler()))
	ociServeMux.Handle(prefix+"/manifests/", http.StripPrefix(prefix+"/manifests/", otr.manifestsHandler()))
}

func (ocs *ociClientSuite) TestClient_Parse() {
	client := &oci.Client{}

	for _, data := range []struct {
		expected *netutil.URL
		name     string
		url      string
	}{
		{
			name: "oci scheme",
			url:  "oci://registry.acme.com/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci-archive scheme",
			url:  "oci-archive://registry.acme.com/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "docker scheme",
			url:  "docker://registry.acme.com/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "docker-archive scheme",
			url:  "docker-archive://registry.acme.com/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "no scheme",
			url:  "registry.acme.com/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci scheme with username, port, tag",
			url:  "oci://username@registry.acme.com:12345/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Username: "username",
				Hostname: "registry.acme.com",
				Port:     "12345",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci scheme with username, password, port, tag",
			url:  "oci://username:password@registry.acme.com:12345/example/image:1.2.3",
			expected: &netutil.URL{
				Scheme:   "oci",
				Username: "username",
				Password: "password",
				Hostname: "registry.acme.com",
				Port:     "12345",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci scheme with username, port, digest",
			url:  "oci://username@registry.acme.com:12345/example/image@" + testSHA,
			expected: &netutil.URL{
				Scheme:   "oci",
				Username: "username",
				Hostname: "registry.acme.com",
				Port:     "12345",
				Path:     "example/image",
				Digest:   testSHA,
			},
		},
		{
			name: "oci scheme with username, password, port, digest",
			url:  "oci://username:password@registry.acme.com:12345/example/image@" + testSHA,
			expected: &netutil.URL{
				Scheme:   "oci",
				Username: "username",
				Password: "password",
				Hostname: "registry.acme.com",
				Port:     "12345",
				Path:     "example/image",
				Digest:   testSHA,
			},
		},
		{
			name:     "git SCP-like form",
			url:      "username:password@github.com:bomctl/bomctl.git",
			expected: nil,
		},
		{
			name:     "missing tag and digest",
			url:      "oci://username:password@registry.acme.com/example/image",
			expected: nil,
		},
	} {
		ocs.Run(data.name, func() {
			actual := client.Parse(data.url)
			ocs.Require().Equal(data.expected, actual, data.url)
		})
	}
}

func TestOCIClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ociClientSuite))
}
