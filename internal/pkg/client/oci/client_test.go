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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

const testSHA string = "sha256:abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF0123456789"

type ociClientSuite struct {
	suite.Suite
}

func (ocs *ociClientSuite) TestClient_Parse() {
	client := &oci.Client{}

	for _, data := range []struct {
		expected *url.ParsedURL
		name     string
		url      string
	}{
		{
			name: "oci scheme",
			url:  "oci://registry.acme.com/example/image:1.2.3",
			expected: &url.ParsedURL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci-archive scheme",
			url:  "oci-archive://registry.acme.com/example/image:1.2.3",
			expected: &url.ParsedURL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "docker scheme",
			url:  "docker://registry.acme.com/example/image:1.2.3",
			expected: &url.ParsedURL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "docker-archive scheme",
			url:  "docker-archive://registry.acme.com/example/image:1.2.3",
			expected: &url.ParsedURL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "no scheme",
			url:  "registry.acme.com/example/image:1.2.3",
			expected: &url.ParsedURL{
				Scheme:   "oci",
				Hostname: "registry.acme.com",
				Path:     "example/image",
				Tag:      "1.2.3",
			},
		},
		{
			name: "oci scheme with username, port, tag",
			url:  "oci://username@registry.acme.com:12345/example/image:1.2.3",
			expected: &url.ParsedURL{
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
			expected: &url.ParsedURL{
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
			expected: &url.ParsedURL{
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
			expected: &url.ParsedURL{
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
