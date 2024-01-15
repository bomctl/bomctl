// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: pkg/utils/utils.go
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
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/bom-squad/protobom/pkg/reader"
	"github.com/bom-squad/protobom/pkg/sbom"
)

var errUnsupportedURL = errors.New("unsupported URL scheme")

func getBOMReferences(document *sbom.Document) (refs []*sbom.ExternalReference) {
	for _, node := range document.NodeList.Nodes {
		for _, ref := range node.GetExternalReferences() {
			if ref.Type == sbom.ExternalReference_BOM {
				refs = append(refs, ref)
			}
		}
	}

	return
}

func parseSBOMData(data []byte) (document *sbom.Document, err error) {
	sbomReader := reader.New()
	bytesReader := bytes.NewReader(data)
	document, err = sbomReader.ParseStream(bytesReader)

	return
}

func DownloadHTTP(url, outputFile string, auth *basicAuthCredentials) (*sbom.Document, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if auth != nil {
		req.Header.Add("Authorization", "Basic "+basicAuth(auth.username, auth.password))
	}

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
	if outputFile != "" {
		out, err := os.Create(outputFile)
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

	document, err := parseSBOMData(respBytes)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return document, nil
}

func FetchSBOM(url, outputFile string) (err error) {
	var document *sbom.Document
	parsedURL := ParseURL(url)

	switch parsedURL.Scheme {
	case "git":
	case "http", "https":
		document, err = DownloadHTTP(url, outputFile, nil)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	case "oci":
	default:
		return fmt.Errorf("%w: %s", errUnsupportedURL, parsedURL.Scheme)
	}

	// Fetch externally referenced BOMs
	var idx uint8
	for _, ref := range getBOMReferences(document) {
		if outputFile != "" {
			// Matches base filename, excluding extension
			baseFilename := regexp.MustCompile(`^([^\.]+)?`).FindString(filepath.Base(outputFile))

			outputFile = fmt.Sprintf("%s-%d.%s",
				filepath.Join(filepath.Dir(outputFile), baseFilename),
				idx,
				filepath.Ext(outputFile),
			)
		}

		err := FetchSBOM(ref.Url, outputFile)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
