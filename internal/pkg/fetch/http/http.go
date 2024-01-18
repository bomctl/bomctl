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
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bom-squad/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/url"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

var client = http.DefaultClient

type HTTPFetcher struct {
	OutputFile string
}

func (hf *HTTPFetcher) Fetch(parsedURL *url.ParsedURL, auth *url.BasicAuth) (*sbom.Document, error) {
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
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
	if hf.OutputFile != "" {
		out, err := os.Create(hf.OutputFile)
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
