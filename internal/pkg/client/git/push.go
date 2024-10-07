// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/push.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -----------------------------------------------------------------------------
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
// -----------------------------------------------------------------------------

package git

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) AddFile(pushURL, id string, opts *options.PushOptions) error {
	document, err := getDocument(id, opts.Options)
	if err != nil {
		return err
	}

	url := client.Parse(pushURL)
	name := url.Fragment

	// Create any parent directories specified in fragment.
	if dir := filepath.Dir(name); dir != "." {
		if err := client.worktree.Filesystem.MkdirAll(dir, fs.ModePerm); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	file, err := client.worktree.Filesystem.Create(name)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer file.Close()

	opts.Logger.Info("Writing document", "name", name)

	// Write the file specified in the URL fragment.
	if err := writer.New(writer.WithFormat(opts.Format)).WriteStream(document, file); err != nil {
		return fmt.Errorf("failed to write file %s: %w", name, err)
	}

	// Stage SBOM file to index.
	if _, err := client.worktree.Add(name); err != nil {
		return fmt.Errorf("failed to stage file %s for commit: %w", name, err)
	}

	return nil
}

func (client *Client) PreparePush(pushURL string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	auth := &netutil.BasicAuth{Username: url.Username, Password: url.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return fmt.Errorf("setting .netrc auth: %w", err)
		}
	}

	// Clone the repository into memory.
	return client.cloneRepo(url, auth, opts.Options)
}

func (client *Client) Push(pushURL string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	auth := &netutil.BasicAuth{Username: url.Username, Password: url.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return fmt.Errorf("failed to set auth: %w", err)
		}
	}

	author := &object.Signature{
		Name:  "bomctl",
		Email: "bomctl@users.noreply.github.com",
		When:  time.Now(),
	}

	// Commit written SBOM file to cloned repo.
	if _, err := client.worktree.Commit(
		fmt.Sprintf("bomctl push of %s", url.Fragment), &git.CommitOptions{All: true, Author: author},
	); err != nil {
		return fmt.Errorf("committing worktree: %w", err)
	}

	// Push changes to remote repository.
	if err := client.repo.Push(&git.PushOptions{Auth: auth}); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("pushing to remote %s: %w", url, err)
		}

		opts.Logger.Warn("Already up-to-date; no changes pushed to remote")
	}

	return nil
}

func getDocument(sbomID string, opts *options.Options) (*sbom.Document, error) {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	opts.Logger.Debug("Retrieving document", "id", sbomID)

	// Retrieve SBOM document from database.
	doc, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return doc, nil
}
