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
	"io"
	neturl "net/url"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) PrepareFetch(url *neturl.URL, auth *netutil.BasicAuth, opts *options.Options) error {
	return client.cloneRepo(url, auth, opts)
}

func (client *Client) Fetch(fetchURL string, _opts *options.FetchOptions) ([]byte, error) {
	url, err := neturl.Parse(fetchURL)
	if err != nil {
		return nil, fmt.Errorf("parsing url %s: %w", url.String(), err)
	}

	// Read the file specified in the URL fragment.
	file, err := client.worktree.Filesystem.Open(url.Fragment)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", url.Fragment, err)
	}

	defer file.Close()

	sbomData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", url.Fragment, err)
	}

	return sbomData, nil
}
