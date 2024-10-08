// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/http/fetch.go
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

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

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

	req, err := http.NewRequestWithContext(context.Background(), "GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request to %s: %w", url.String(), err)
	}

	auth.SetAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed request to %s: %w", url.String(), err)
	}

	defer resp.Body.Close()

	sbomData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return sbomData, nil
}
