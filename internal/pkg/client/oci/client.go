// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/oci/fetch.go
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
	"errors"
	"fmt"
	"regexp"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errMultipleSBOMs = errors.New("more than one SBOM document identified in OCI image")

type Client struct{}

func (*Client) Name() string {
	return "OCI"
}

func (*Client) RegExp() *regexp.Regexp {
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

func (client *Client) Parse(fetchURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := client.RegExp()
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
