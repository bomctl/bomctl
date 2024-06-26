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

	"github.com/charmbracelet/log"
	"github.com/jdx/go-netrc"
	"github.com/protobom/protobom/pkg/reader"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/fetch/git"
	"github.com/bomctl/bomctl/internal/pkg/fetch/http"
	"github.com/bomctl/bomctl/internal/pkg/fetch/oci"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errUnsupportedURL = errors.New("failed to parse URL; see `bomctl fetch --help` for valid URL patterns")

type (
	Fetcher interface {
		url.Parser
		Fetch(*url.ParsedURL, *url.BasicAuth) ([]byte, error)
		Name() string
	}

	FetchOptions struct {
		Logger     *log.Logger
		OutputFile *os.File
		CacheDir   string
		ConfigFile string
		Debug      bool
		UseNetRC   bool
	}
)

func Fetch(sbomURL string, opts *FetchOptions) error {
	document, err := fetchDocument(sbomURL, opts)
	if err != nil {
		return err
	}

	backend := db.NewBackend()
	backend.Options.DatabaseFile = filepath.Join(opts.CacheDir, db.DatabaseFile)
	backend.Options.Debug = opts.Debug

	if err := backend.InitClient(); err != nil {
		return fmt.Errorf("failed to initialize backend client: %w", err)
	}

	defer backend.CloseClient()

	// Insert fetched document data into database.
	if err := backend.AddDocument(document); err != nil {
		return fmt.Errorf("error adding document: %w", err)
	}

	if opts.OutputFile == nil || opts.OutputFile.Name() == "" {
		return nil
	}

	// Fetch externally referenced BOMs
	var idx uint8

	extRefs, err := backend.GetExternalReferencesByDocumentID(document.Metadata.Id, "BOM")
	if err != nil {
		return fmt.Errorf("error getting external references: %w", err)
	}

	for _, ref := range extRefs {
		idx++

		refOutput, err := getRefFile(opts.OutputFile)
		if err != nil {
			return err
		}

		defer refOutput.Close()

		if err := Fetch(ref.Url, opts); err != nil {
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

func fetchDocument(sbomURL string, opts *FetchOptions) (*sbom.Document, error) {
	fetcher, err := NewFetcher(sbomURL)
	if err != nil {
		return nil, err
	}

	opts.Logger.Info(fmt.Sprintf("Fetching from %s URL", fetcher.Name()), "url", sbomURL)

	parsedURL := fetcher.Parse(sbomURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := setNetRCAuth(parsedURL.Hostname, auth); err != nil {
			return nil, err
		}
	}

	sbomData, err := fetcher.Fetch(parsedURL, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", sbomURL, err)
	}

	if opts.OutputFile != nil {
		// Write the SBOM document bytes to file.
		if _, err = io.Copy(opts.OutputFile, bytes.NewReader(sbomData)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", opts.OutputFile.Name(), err)
		}
	}

	sbomReader := reader.New()

	document, err := sbomReader.ParseStream(bytes.NewReader(sbomData))
	if err != nil {
		return nil, fmt.Errorf("error parsing SBOM file content: %w", err)
	}

	return document, nil
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
