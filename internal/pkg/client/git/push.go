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
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/protobom/protobom/pkg/sbom"

	"github.com/bomctl/bomctl/internal/pkg/url"
)

func (*Client) Push(_document *sbom.Document, _parsedURL *url.ParsedURL, _auth *url.BasicAuth) error {
	// Clone the repository into the temp directory
	repo, tmpDir, err := CloneRepo(_parsedURL, _auth)
	if err != nil {
		return fmt.Errorf("failed to clone Git repository: %w", err)
	}

	repoTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", _parsedURL.Path, err)
	}

	filePath := filepath.Join(tmpDir, _parsedURL.Fragment)

	// Write the file specified in the URL fragment
	err = os.WriteFile(filePath, []byte(_document.String()), fs.ModePerm) // fs.ModePerm == 0777
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", _parsedURL.Fragment, err)
	}

	// Commit written SBOM file to cloned repo
	_, err = repoTree.Commit(fmt.Sprintf("bomctl push of %s", filePath), &git.CommitOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to commit file %s: %w", filePath, err)
	}

	// Push changes to Repo remote
	err = repo.Push(&git.PushOptions{Auth: _auth})
	if err != nil {
		return fmt.Errorf("failed to push to remote %s: %w", _parsedURL, err)
	}

	return nil
}
