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
	"os"
	"path/filepath"
	"regexp"

	"github.com/bom-squad/protobom/pkg/sbom"
	"github.com/jdx/go-netrc"

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

func Exec(sbomURL, outputFile string, useNetRC bool) error {
	var fetcher Fetcher
	logger := utils.NewLogger("fetch")

	switch {
	case (&oci.OCIFetcher{}).Parse(sbomURL) != nil:
		logger.Info("Fetching from OCI URL", "url", sbomURL)
		fetcher = &oci.OCIFetcher{}
	case (&git.GitFetcher{}).Parse(sbomURL) != nil:
		logger.Info("Fetching from Git URL", "url", sbomURL)
		fetcher = &git.GitFetcher{}
	case (&http.HTTPFetcher{}).Parse(sbomURL) != nil:
		logger.Info("Fetching from HTTP URL", "url", sbomURL)
		fetcher = &http.HTTPFetcher{OutputFile: outputFile}
	default:
		return fmt.Errorf("%w", errUnsupportedURL)
	}

	parsedURL := fetcher.Parse(sbomURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if useNetRC {
		if err := setNetRCAuth(parsedURL.Hostname, auth); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

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

		err := Exec(ref.Url, outputFile, useNetRC)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func setNetRCAuth(hostname string, auth *url.BasicAuth) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	authFile, err := netrc.Parse(filepath.Join(home, ".netrc"))
	if err != nil {
		return fmt.Errorf("failed to parse .netrc file: %w", err)
	}

	// Use credentials in .netrc if entry for the hostname is found
	if machine := authFile.Machine(hostname); machine != nil {
		auth.Username = machine.Get("login")
		auth.Password = machine.Get("password")
	}

	return nil
}
