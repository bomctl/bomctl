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
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/sbom"
	protowriter "github.com/protobom/protobom/pkg/writer"

	"github.com/bomctl/bomctl/internal/pkg/db"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

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

	// Create temp directory to clone into.
	tmpDir, err := os.MkdirTemp(os.TempDir(), strings.ReplaceAll(parsedURL.Path, "/", "-"))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone the repository into the temp directory
	repo, err := cloneRepo(tmpDir, parsedURL, auth, opts.Options)
	if err != nil {
		return fmt.Errorf("failed to clone Git repository: %w", err)
	}

	// Defer removing the temp directory with the cloned repo, to clean up
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, parsedURL.Fragment)

	err = addFile(repo, filePath, opts, doc, parsedURL)
	if err != nil {
		return fmt.Errorf("failed to commit file %s: %w", filePath, err)
	}

	// Push changes to Repo remote
	err = pushFile(repo, parsedURL, auth)
	if err != nil {
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

func addFile(repo *git.Repository, filePath string, opts *options.PushOptions,
	doc *sbom.Document, parsedURL *url.ParsedURL,
) error {
	writer := protowriter.New(protowriter.WithFormat(opts.Format))

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", parsedURL.Fragment, err)
	}

	opts.Logger.Debug("Creating any needed directories prior to creating file")

	if _, err := os.Stat(path.Dir(filePath)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(filePath), fs.ModePerm) // fs.ModePerm == 0777
		if err != nil {
			return fmt.Errorf("failed to create required parent directory %s: %w", parsedURL.Fragment, err)
		}
	}

	opts.Logger.Debug("Writing document to: %s", parsedURL.Fragment)
	// Write the file specified in the URL fragment
	err = writer.WriteFile(doc, filePath)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", parsedURL.Fragment, err)
	}

	// Stage written sbom for addition
	_, err = worktree.Add(parsedURL.Fragment)
	if err != nil {
		return fmt.Errorf("failed to stage file %s for commit: %w", parsedURL.Fragment, err)
	}

	// Commit written SBOM file to cloned repo
	_, err = worktree.Commit(fmt.Sprintf("bomctl push of %s", filePath), &git.CommitOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to commit file %s: %w", filePath, err)
	}

	return nil
}

func getDocument(sbomID string, opts *options.Options) (*sbom.Document, error) {
	backend, err := db.BackendFromContext(opts.Context())
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	opts.Logger.Debug("Pulling document: %s", sbomID)
	// Retrieve SBOM document from database.
	doc, err := backend.GetDocumentByID(sbomID)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return doc, nil
}
