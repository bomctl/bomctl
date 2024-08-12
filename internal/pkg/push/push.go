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
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdx/go-netrc"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/client"
	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

func Push(sbomIDs []string, destPath string, opts *client.PushOptions) error {
	backend := db.NewBackend().
		Debug(opts.Debug).
		WithDatabaseFile(filepath.Join(opts.CacheDir, db.DatabaseFile)).
		WithLogger(opts.Logger)

	if err := backend.InitClient(); err != nil {
		return fmt.Errorf("failed to initialize backend client: %w", err)
	}

	defer backend.CloseClient()

	for _, id := range sbomIDs {
		// retrieve sbom document from database.
		document, err := backend.GetDocumentByID(id)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = pushDocument(document, destPath, opts)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func pushDocument(document *sbom.Document, destPath string, opts *client.PushOptions) error {
	opts.Logger.Info("Pushing Document", "sbomID", document.Metadata.Id)

	pusher, err := client.New(destPath)
	if err != nil {
		return fmt.Errorf("creating push client: %w", err)
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
