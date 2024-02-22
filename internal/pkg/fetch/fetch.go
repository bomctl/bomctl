// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/fetch/fetch.go
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
package fetch

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/bom-squad/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/fetch/git"
	"github.com/bomctl/bomctl/internal/pkg/fetch/http"
	"github.com/bomctl/bomctl/internal/pkg/fetch/oci"
	"github.com/bomctl/bomctl/internal/pkg/url"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

var errUnsupportedURL = errors.New("unsupported URL scheme")

type Fetcher interface {
	url.URLParser
	Fetch(*url.ParsedURL, *url.BasicAuth) (*sbom.Document, error)
}

func Exec(sbomURL, outputFile string) error {
	var fetcher Fetcher

	switch {
	case (&oci.OCIFetcher{}).Parse(sbomURL) != nil:
		fetcher = &oci.OCIFetcher{}
	case (&git.GitFetcher{}).Parse(sbomURL) != nil:
		fetcher = &git.GitFetcher{}
	case (&http.HTTPFetcher{}).Parse(sbomURL) != nil:
		fetcher = &http.HTTPFetcher{OutputFile: outputFile}
	default:
		return fmt.Errorf("%w", errUnsupportedURL)
	}

	parsedURL := fetcher.Parse(sbomURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}
	document, err := fetcher.Fetch(parsedURL, auth)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Insert fetched document data into database.
	err = db.AddDocument(document)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Fetch externally referenced BOMs
	var idx uint8
	for _, ref := range utils.GetBOMReferences(document) {
		idx += 1

		if outputFile != "" {
			// Matches base filename, excluding extension
			baseFilename := regexp.MustCompile(`^([^\.]+)?`).FindString(filepath.Base(outputFile))

			outputFile = fmt.Sprintf("%s-%d.%s",
				filepath.Join(filepath.Dir(outputFile), baseFilename),
				idx,
				filepath.Ext(outputFile),
			)
		}

		err := Exec(ref.Url, outputFile)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
