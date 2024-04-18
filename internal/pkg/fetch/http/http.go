// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
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
	"os"
	"regexp"

	"github.com/bom-squad/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/url"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

var client = http.DefaultClient

type Fetcher struct {
	OutputFile string
}

func (fetcher *Fetcher) Name() string {
	return "HTTP"
}

func (fetcher *Fetcher) RegExp() *regexp.Regexp {
	return regexp.MustCompile(
		fmt.Sprintf("%s%s%s%s",
			`((?P<scheme>https?)://)`,
			`((?P<username>[^:]+)(?::(?P<password>[^@]+))?(?:@))?`,
			`(?P<hostname>[^@/?#:]*)(?::(?P<port>\d+)?)?`,
			`(/?(?P<path>[^@?#]*))(\?(?P<query>[^#]*))?(#(?P<fragment>.*))?`,
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

func (fetcher *Fetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) (*sbom.Document, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	auth.SetAuth(req)

	// Get the data
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	// Create the file if specified at the command line
	if fetcher.OutputFile != "" {
		out, err := os.Create(fetcher.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		defer out.Close()

		// Write the response body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	document, err := utils.ParseSBOMData(respBytes)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return document, nil
}
