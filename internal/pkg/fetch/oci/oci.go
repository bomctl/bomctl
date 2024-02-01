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
	"regexp"

	"github.com/bom-squad/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

type OCIFetcher struct{}

func (of *OCIFetcher) RegExp() *regexp.Regexp {
	return regexp.MustCompile(`((?P<scheme>oci)(?:://))?(?P<hostname>[^/:]*)[/:]?(?P<path>[^:]*)?:?(?P<tag>(.*))`)
}

func (of *OCIFetcher) Parse(fetchURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := of.RegExp()
	match := pattern.FindStringSubmatch(fetchURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	return &url.ParsedURL{
		Scheme:   results["scheme"],
		Hostname: results["hostname"],
		Path:     results["path"],
		Tag:      results["tag"],
	}
}

// Not yet implemented.
func (of *OCIFetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) (*sbom.Document, error) {
	return nil, nil
}
