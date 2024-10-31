// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/git/fetch.go
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
	"fmt"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	url := client.Parse(fetchURL)
	auth := &netutil.BasicAuth{Username: url.Username, Password: url.Password}

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return nil, fmt.Errorf("failed to set auth: %w", err)
		}
	}

	// Clone the repository into the temp directory
	if err := client.cloneRepo(url, auth, opts.Options); err != nil {
		return nil, fmt.Errorf("cloning Git repository: %w", err)
	}

	// Read the file specified in the URL fragment.
	file, err := client.worktree.Filesystem.Open(url.Fragment)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", url.Fragment, err)
	}

	defer file.Close()

	sbomData := []byte{}
	if _, err := file.Read(sbomData); err != nil {
		return nil, fmt.Errorf("reading file %s: %w", url.Fragment, err)
	}

	return sbomData, nil
}
