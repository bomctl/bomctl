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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/jdx/go-netrc"
	"github.com/protobom/protobom/pkg/reader"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/fetch/git"
	"github.com/bomctl/bomctl/internal/pkg/fetch/http"
	"github.com/bomctl/bomctl/internal/pkg/fetch/oci"
	"github.com/bomctl/bomctl/internal/pkg/url"
	"github.com/bomctl/bomctl/internal/pkg/utils"
)

var errUnsupportedURL = errors.New("unsupported URL scheme")

type Fetcher interface {
	url.Parser
	Fetch(*url.ParsedURL, *url.BasicAuth) ([]byte, error)
	Name() string
}

func Fetch(sbomURL string, outputFile *os.File, useNetRC bool) error { //nolint:cyclop,funlen
	logger := utils.NewLogger("fetch")

	fetcher, err := NewFetcher(sbomURL)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("Fetching from %s URL", fetcher.Name()), "url", sbomURL)

	parsedURL := fetcher.Parse(sbomURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if useNetRC {
		if err := setNetRCAuth(parsedURL.Hostname, auth); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	sbomData, err := fetcher.Fetch(parsedURL, auth)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if outputFile != nil {
		// Write the SBOM document bytes to file.
		if _, err = io.Copy(outputFile, bytes.NewReader(sbomData)); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputFile.Name(), err)
		}
	}

	sbomReader := reader.New()

	document, err := sbomReader.ParseStream(bytes.NewReader(sbomData))
	if err != nil {
		return fmt.Errorf("error parsing SBOM file content: %w", err)
	}

	// Insert fetched document data into database.
	err = db.AddDocument(document)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if outputFile == nil || outputFile.Name() == "" {
		return nil
	}

	// Fetch externally referenced BOMs
	var idx uint8
	for _, ref := range utils.GetBOMReferences(document) {
		idx++

		refOutput, err := getRefFile(outputFile)
		if err != nil {
			return err
		}

		defer refOutput.Close()

		if err := Fetch(ref.Url, refOutput, useNetRC); err != nil {
			return err
		}
	}

	return nil
}

func NewFetcher(sbomURL string) (Fetcher, error) {
	for _, fetcher := range []Fetcher{&git.Fetcher{}, &http.Fetcher{}, &oci.Fetcher{}} {
		if parsedURL := fetcher.Parse(sbomURL); parsedURL != nil {
			return fetcher, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", errUnsupportedURL, sbomURL)
}

func getRefFile(parentFile *os.File) (*os.File, error) {
	idx := 0

	// Matches base filename, excluding extension
	baseFilename := regexp.MustCompile(`^([^\.]+)?`).FindString(filepath.Base(parentFile.Name()))

	suffix := regexp.MustCompile(`^.*-(\d+)`).FindString(baseFilename)

	if suffix != "" {
		var err error

		idx, err = strconv.Atoi(suffix)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	idx++

	outputFile := fmt.Sprintf("%s-%d.%s",
		filepath.Join(filepath.Dir(parentFile.Name()), baseFilename),
		idx,
		filepath.Ext(parentFile.Name()),
	)

	refOutput, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return refOutput, nil
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
