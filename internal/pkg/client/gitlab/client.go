// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/gitlab/client.go
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

package gitlab

import (
	"fmt"
	"regexp"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
)

type Client struct {
	ProjectProvider
	BranchProvider
	CommitProvider
	DependencyListExporter
	Export      *gitlab.DependencyListExport
	GitLabToken string
}

func (*Client) Name() string {
	return "GitLab"
}

func (*Client) RegExp() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("(?i)^%s%s%s$",
		`(?P<scheme>https?|git|ssh):\/\/`,
		`(?P<hostname>[^@\/?#:]+gitlab[^@\/?#:]+)(?::(?P<port>\d+))?/`,
		`(?P<path>[^@#]+)@(?P<gitRef>\S+)`))
}

func (client *Client) Parse(rawURL string) *netutil.URL {
	results := map[string]string{}
	pattern := client.RegExp()
	match := pattern.FindStringSubmatch(rawURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname", "path", "gitRef"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	return &netutil.URL{
		Scheme:   results["scheme"],
		Hostname: results["hostname"],
		Port:     results["port"],
		Path:     results["path"],
		GitRef:   results["gitRef"],
	}
}
