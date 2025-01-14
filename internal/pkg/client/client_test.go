// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/client_test.go
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

package client_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const (
	manifestTag = "v1"
	repoName    = "oci-client-test"
	testSHA     = "sha256:abcdef0123456789ABCDEF0123456789abcdef0123456789ABCDEF0123456789"
)

type clientSuite struct {
	suite.Suite
	*options.Options
}

// URL: [scheme:][//[userinfo@]host][/]path[?query][#fragment]
//
// https://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json
//
// scheme: https
// userinfo: [username, password]
// host: github.com:12345 (parse port with url.Port())
// path: bomctl/bomctl.git
// fragment: sbom.cdx.json
// Query: ref=main (parse query with url.Query() and get value with query.Get("ref"))

func (cs *clientSuite) TestClient_DetermineClient() {
	for _, data := range []struct {
		expected string
		name     string
		url      string
		err      bool
	}{
		{
			name:     "https scheme",
			url:      "https://another.github.com/bomctl/bomctl",
			expected: "github",
			err:      false,
		},
		{
			name:     "oci scheme",
			url:      "oci://registry.acme.com/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "oci-archive scheme",
			url:      "oci-archive://registry.acme.com/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "docker scheme",
			url:      "docker://registry.acme.com/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "docker-archive scheme",
			url:      "docker-archive://registry.acme.com/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "oci scheme with username, port, tag",
			url:      "oci://username@registry.acme.com:12345/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "oci scheme with username, password, port, tag",
			url:      "oci://username:password@registry.acme.com:12345/example/image?ref=1.2.3",
			expected: "OCI",
			err:      false,
		},
		{
			name:     "oci scheme with username, port, digest",
			url:      "oci://username@registry.acme.com:12345/example/image?ref=" + testSHA,
			expected: "OCI",
			err:      false,
		},
		{
			name:     "oci scheme with username, password, port, digest",
			url:      "oci://username:password@registry.acme.com:12345/example/image?ref=" + testSHA,
			expected: "OCI",
			err:      false,
		},
		{
			name:     "git SCP-like form",
			url:      "username:password@github.com:bomctl/bomctl.git",
			expected: "",
			err:      true,
		},
		{
			name:     "missing ref",
			url:      "oci://username:password@registry.acme.com/example/image",
			expected: "",
			err:      true,
		},
		{
			name:     "http scheme",
			url:      "http://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "https scheme with username, port",
			url:      "https://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "https scheme with username, password, port",
			url:      "https://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "https scheme with username",
			url:      "https://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "ssh scheme",
			url:      "ssh://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "ssh scheme with username, port",
			url:      "ssh://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "ssh scheme with username, password, port",
			url:      "ssh://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "ssh scheme with username",
			url:      "ssh://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "git scheme",
			url:      "git://github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "git scheme with username, port",
			url:      "git://git@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "git scheme with username, password, port",
			url:      "git://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "git scheme with username",
			url:      "git://git@github.com/bomctl/bomctl.git?ref=main#sbom.cdx.json",
			expected: "Git",
		},
		{
			name:     "path does not end in .git",
			url:      "https://github.com/bomctl/bomctl?ref=main#sbom.cdx.json",
			expected: "",
			err:      true,
		},
		{
			name:     "missing git ref",
			url:      "https://github.com/bomctl/bomctl.git#sbom.cdx.json",
			expected: "",
			err:      true,
		},
		{
			name:     "missing path to SBOM file",
			url:      "https://github.com/bomctl/bomctl.git?ref=main",
			expected: "",
			err:      true,
		},
	} {
		cs.Run(data.name, func() {
			actual, err := client.New(data.url)
			if data.err {
				cs.Require().EqualError(err, errors.ErrUnsupported.Error())
				cs.Require().Nil(actual, data.url)
			} else {
				cs.Require().NoError(err)
				cs.Require().Equal(data.expected, actual.Name(), data.url)
			}
		})
	}
}

func TestClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(clientSuite))
}
