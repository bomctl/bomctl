// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/github/fetch.go
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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"

	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	url := client.Parse(fetchURL)

	if url.Fragment != "" && url.GitRef != "" {
		return client.gitFetch(fetchURL, opts)
	}

	ctx := context.Background()
	auth := netutil.NewBasicAuth(url.Username, url.Password)

	if opts.UseNetRC {
		if err := auth.UseNetRC(url.Hostname); err != nil {
			return nil, fmt.Errorf("failed to set auth: %w", err)
		}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: auth.Password})
	tc := oauth2.NewClient(ctx, ts)
	client.ghClient = *github.NewClient(tc)

	repoURL := strings.Split(url.Path, "/")
	owner := repoURL[0]
	repo := repoURL[1] + "/dependency-graph/sbom"
	u := fmt.Sprintf("repos/%s/%s", owner, repo)

	req, err := client.ghClient.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", ("Bearer " + auth.Password))

	resp, err := client.ghClient.BareDo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data map[string]map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Github API returns the sbom inside of an object named sbom, so need to drill down one layer
	sbomData, err := json.Marshal(data["sbom"])
	if err != nil {
		return nil, fmt.Errorf("failed to extract SBOM from response: %w", err)
	}

	return sbomData, nil
}

func (client *Client) gitFetch(fetchURL string, opts *options.FetchOptions) ([]byte, error) {
	urlParts := strings.Split(fetchURL, "@")
	gitURL := "git+" + urlParts[0] + ".git@" + urlParts[1]

	sbomData, err := client.gitClient.Fetch(gitURL, opts)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return sbomData, nil
}
