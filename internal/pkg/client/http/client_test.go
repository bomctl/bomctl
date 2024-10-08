// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/http/client_test.go
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

package http_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bomctl/bomctl/internal/pkg/client/http"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
)

type httpClientSuite struct {
	suite.Suite
}

func (hcs *httpClientSuite) TestClient_Parse() {
	client := &http.Client{}

	for _, data := range []struct {
		expected *netutil.URL
		name     string
		url      string
	}{
		{
			name:     "HTTP scheme and hostname only",
			url:      "http://example.acme.com",
			expected: &netutil.URL{Scheme: "http", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTPS scheme and hostname only",
			url:      "https://example.acme.com",
			expected: &netutil.URL{Scheme: "https", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTPS username@hostname",
			url:      "https://user@example.acme.com",
			expected: &netutil.URL{Scheme: "https", Username: "user", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTP hostname only",
			url:      "example.acme.com",
			expected: nil,
		},
	} {
		hcs.Run(data.name, func() {
			actual := client.Parse(data.url)
			hcs.Require().Equal(data.expected, actual, data.url)
		})
	}
}

func TestHTTPClientSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(httpClientSuite))
}
