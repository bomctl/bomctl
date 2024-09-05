// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch.go
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
	"os"
	"path/filepath"
	"strings"

	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

func (client *Client) Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	parsedURL := client.Parse(fetchURL)
	auth := &url.BasicAuth{Username: parsedURL.Username, Password: parsedURL.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(parsedURL.Hostname); err != nil {
			return nil, fmt.Errorf("failed to set auth: %w", err)
		}
	}

	// Create temp directory to clone into.
	tmpDir, err := os.MkdirTemp(os.TempDir(), strings.ReplaceAll(parsedURL.Path, "/", "-"))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone the repository into the temp directory
	_, err = cloneRepo(tmpDir, parsedURL, auth, opts.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to clone Git repository: %w", err)
	}

	defer os.RemoveAll(tmpDir)

	// Read the file specified in the URL fragment
	sbomData, err := os.ReadFile(filepath.Join(tmpDir, parsedURL.Fragment))
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", parsedURL.Fragment, err)
	}

	return sbomData, nil
}
