// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/github/client.go
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

package github

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v66/github"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
)

type Client struct {
	ghClient github.Client
}

func (*Client) Name() string {
	return "github"
}

func (*Client) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("^%s%s%s%s$",
			`(?P<scheme>https?|git|ssh):\/\/?`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`(?P<hostname>github(\.[A-Za-z0-9_-]+)*\.com+)(?::(?P<port>\d+))?`,
			`(?:[\/:](?P<path>[^@#]+)@?)`,
		),
	)
}

func (client *Client) Parse(rawURL string) *netutil.URL {
	results := map[string]string{}
	pattern := client.RegExp()
	match := pattern.FindStringSubmatch(rawURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	if results["scheme"] == "" {
		results["scheme"] = "https"
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname", "path"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	const length = 2

	pathComponents := strings.Split(results["path"], "/")

	if len(pathComponents) != length {
		return nil
	}

	return &netutil.URL{
		Scheme:   results["scheme"],
		Username: results["username"],
		Password: results["password"],
		Hostname: results["hostname"],
		Path:     results["path"],
		Port:     results["port"],
	}
}
