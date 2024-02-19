// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/fetch/git/git_test.go
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
package git

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

func TestGitFetcherParse(t *testing.T) {
	fetcher := &GitFetcher{}

	for _, data := range []struct {
		expected *url.ParsedURL
		name     string
		url      string
	}{
		{
			name:     "git+http scheme",
			url:      "git+http://github.com/bomctl/bomctl.git",
			expected: &url.ParsedURL{Scheme: "git", Hostname: "github.com", Path: "bomctl/bomctl.git"},
		},
		{
			name: "git+https scheme with username, port",
			url:  "git+https://git@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "git+https scheme with username, password, port",
			url:  "git+https://username:password@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "git+https scheme with ref, file path",
			url:  "git+https://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git+https scheme with username, ref, file path",
			url:  "git+https://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name:     "ssh scheme",
			url:      "ssh://github.com/bomctl/bomctl.git",
			expected: &url.ParsedURL{Scheme: "ssh", Hostname: "github.com", Path: "bomctl/bomctl.git"},
		},
		{
			name: "ssh scheme with username, port",
			url:  "ssh://git@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "ssh scheme with username, password, port",
			url:  "ssh://username:password@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "ssh scheme with ref, file path",
			url:  "ssh://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "ssh scheme with username, ref, file path",
			url:  "ssh://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name:     "git scheme",
			url:      "git://github.com/bomctl/bomctl.git",
			expected: &url.ParsedURL{Scheme: "git", Hostname: "github.com", Path: "bomctl/bomctl.git"},
		},
		{
			name: "git scheme with username, port",
			url:  "git://git@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "git scheme with username, password, port",
			url:  "git://username:password@github.com:12345/bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Port:     "12345",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "git scheme with ref, file path",
			url:  "git://github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git scheme with username, ref, file path",
			url:  "git://git@github.com/bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "git",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name:     "git SCP-like syntax",
			url:      "github.com:bomctl/bomctl.git",
			expected: &url.ParsedURL{Hostname: "github.com", Path: "bomctl/bomctl.git"},
		},
		{
			name:     "git SCP-like syntax with username",
			url:      "git@github.com:bomctl/bomctl.git",
			expected: &url.ParsedURL{Scheme: "ssh", Username: "git", Hostname: "github.com", Path: "bomctl/bomctl.git"},
		},
		{
			name: "git SCP-like syntax with username, password",
			url:  "username:password@github.com:bomctl/bomctl.git",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "username",
				Password: "password",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
			},
		},
		{
			name: "git SCP-like syntax with ref, file path",
			url:  "github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
		{
			name: "git SCP-like syntax with username, ref, file path",
			url:  "git@github.com:bomctl/bomctl.git@main#sbom.cdx.json",
			expected: &url.ParsedURL{
				Scheme:   "ssh",
				Username: "git",
				Hostname: "github.com",
				Path:     "bomctl/bomctl.git",
				GitRef:   "main",
				Fragment: "sbom.cdx.json",
			},
		},
	} {
		t.Run(data.name, func(t *testing.T) {
			actual := fetcher.Parse(data.url)

			assert.Equal(t, data.expected, actual, data.url)
		})
	}
}
