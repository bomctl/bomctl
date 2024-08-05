// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/fetch/http/http.go
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
package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

type Fetcher struct{}

func (*Fetcher) Name() string {
	return "HTTP"
}

func (*Fetcher) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("%s%s%s%s",
			`((?P<scheme>https?):\/\/)`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`(?P<hostname>[^@\/?#:]+)(?::(?P<port>\d+))?`,
			`(\/?(?P<path>[^@?#]*))(\?(?P<query>[^#]*))?(#(?P<fragment>.*))?`,
		),
	)
}

func (fetcher *Fetcher) Parse(fetchURL string) *url.ParsedURL {
	results := map[string]string{}
	pattern := fetcher.RegExp()
	match := pattern.FindStringSubmatch(fetchURL)

	for idx, name := range match {
		results[pattern.SubexpNames()[idx]] = name
	}

	// Ensure required map fields are present.
	for _, required := range []string{"scheme", "hostname"} {
		if value, ok := results[required]; !ok || value == "" {
			return nil
		}
	}

	return &url.ParsedURL{
		Scheme:   results["scheme"],
		Username: results["username"],
		Password: results["password"],
		Hostname: results["hostname"],
		Port:     results["port"],
		Path:     results["path"],
		Query:    results["query"],
		Fragment: results["fragment"],
	}
}

func (*Fetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request to %s: %w", parsedURL.String(), err)
	}

	auth.SetAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed request to %s: %w", parsedURL.String(), err)
	}

	defer resp.Body.Close()

	sbomData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return sbomData, nil
}
