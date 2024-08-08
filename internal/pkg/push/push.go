// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/push/push.go
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
package push

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdx/go-netrc"
	"github.com/protobom/protobom/pkg/formats"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errUnsupportedURL = errors.New("failed to parse URL; see `bomctl push --help` for valid URL patterns")

type (
	Pusher interface {
		url.Parser
		Push(*sbom.Document, *url.ParsedURL, *url.BasicAuth) error
		Name() string
	}

	Options struct {
		*options.Options
		Format   formats.Format
		UseNetRC bool
		UseTree  bool
	}
)

func Push(sbomID, destPath string, opts *Options) error {
	backend := db.NewBackend().
		Debug(opts.Debug).
		WithDatabaseFile(filepath.Join(opts.CacheDir, db.DatabaseFile)).
		WithLogger(opts.Logger)

	if err := backend.InitClient(); err != nil {
		return fmt.Errorf("failed to initialize backend client: %w", err)
	}

	defer backend.CloseClient()

	// Insert pushed document data into database.
	document, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Push externally referenced BOMs
	return pushDocument(document, destPath, opts)
}

func New(sbomURL string) (Pusher, error) {
	for _, pusher := range []Pusher{} {
		if parsedURL := pusher.Parse(sbomURL); parsedURL != nil {
			return pusher, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", errUnsupportedURL, sbomURL)
}

func pushDocument(document *sbom.Document, destPath string, opts *Options) error {
	opts.Logger.Info("Pushing Document", "sbomID", document.Metadata.Id)

	pusher, err := NewPusher(destPath)
	if err != nil {
		return err
	}

	opts.Logger.Info(fmt.Sprintf("Pushing to %s URL", pusher.Name()), "url", destPath)

	parsedURL := pusher.Parse(destPath)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := setNetRCAuth(parsedURL.Hostname, auth); err != nil {
			return err
		}
	}

	err = pusher.Push(document, parsedURL, auth)
	if err != nil {
		return fmt.Errorf("failed to push from %s: %w", parsedURL, err)
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
