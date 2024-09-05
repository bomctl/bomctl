// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/push.go
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
package git

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/sbom"
	"github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errAssertReadWriterType = errors.New("type assertion failed")

func (client *Client) AddFile(rw io.ReadWriter, doc *sbom.Document, opts *options.PushOptions) error {
	file, ok := rw.(*os.File)
	if !ok {
		return fmt.Errorf("%w", errAssertReadWriterType)
	}

	relPath, err := filepath.Rel(client.tmpDir, file.Name())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	wt, err := client.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", relPath, err)
	}

	opts.Logger.Debug("Writing document to: %s", relPath)

	// Write the file specified in the URL fragment
	if err := writer.New(writer.WithFormat(opts.Format)).WriteFile(doc, relPath); err != nil {
		return fmt.Errorf("failed to write file %s: %w", relPath, err)
	}

	// Stage written sbom for addition
	_, err = wt.Add(relPath)
	if err != nil {
		return fmt.Errorf("failed to stage file %s for commit: %w", relPath, err)
	}

	// Commit written SBOM file to cloned repo
	_, err = wt.Commit(fmt.Sprintf("bomctl push of %s", relPath), &git.CommitOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to commit file %s: %w", relPath, err)
	}

	return nil
}

func (client *Client) Push(id, pushURL string, opts *options.PushOptions) error {
	doc, err := getDocument(id, opts.Options)
	if err != nil {
		return fmt.Errorf("failed to initialize backend client: %w", err)
	}

	parsedURL := client.Parse(pushURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(parsedURL.Hostname); err != nil {
			return fmt.Errorf("failed to set auth: %w", err)
		}
	}

	// Clone the repository into the temp directory
	if err := client.cloneRepo(parsedURL, auth, opts.Options); err != nil {
		return fmt.Errorf("failed to clone Git repository: %w", err)
	}

	// Defer removing the temp directory with the cloned repo, to clean up
	defer client.removeTmpDir()

	filePath := filepath.Join(client.tmpDir, parsedURL.Fragment)

	if err := os.MkdirAll(filepath.Dir(filePath), fs.ModeAppend); err != nil {
		return fmt.Errorf("failed to create required parent directory %s: %w", parsedURL.Fragment, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := client.AddFile(file, doc, opts); err != nil {
		return fmt.Errorf("failed to commit file %s: %w", filePath, err)
	}

	// Push changes to Repo remote
	if err := pushFile(client.repo, parsedURL, auth); err != nil {
		return fmt.Errorf("failed to push to remote %s: %w", parsedURL, err)
	}

	return nil
}

func pushFile(repo *git.Repository, parsedURL *url.ParsedURL, auth *url.BasicAuth) error {
	err := repo.Push(&git.PushOptions{Auth: auth})
	if err != nil {
		return fmt.Errorf("failed to push to remote %s: %w", parsedURL, err)
	}

	return nil
}

func getDocument(id string, opts *options.Options) (*sbom.Document, error) {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	opts.Logger.Debug("Pulling document: %s", id)
	// Retrieve SBOM document from database.
	doc, err := backend.GetDocumentByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return doc, nil
}
