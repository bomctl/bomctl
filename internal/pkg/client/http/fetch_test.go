// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/http/fetch_test.go
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

	"github.com/stretchr/testify/assert"

	"github.com/bomctl/bomctl/internal/pkg/client/http"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

func TestFetcher_Parse(t *testing.T) {
	t.Parallel()

	fetcher := &http.Client{}

	for _, data := range []struct {
		expected *url.ParsedURL
		name     string
		url      string
	}{
		{
			name:     "HTTP scheme and hostname only",
			url:      "http://example.acme.com",
			expected: &url.ParsedURL{Scheme: "http", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTPS scheme and hostname only",
			url:      "https://example.acme.com",
			expected: &url.ParsedURL{Scheme: "https", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTPS username@hostname",
			url:      "https://user@example.acme.com",
			expected: &url.ParsedURL{Scheme: "https", Username: "user", Hostname: "example.acme.com"},
		},
		{
			name:     "HTTP hostname only",
			url:      "example.acme.com",
			expected: nil,
		},
	} {
		t.Run(data.name, func(t *testing.T) {
			t.Parallel()

			actual := fetcher.Parse(data.url)

			assert.Equal(t, data.expected, actual, data.url)
		})
	}
}
