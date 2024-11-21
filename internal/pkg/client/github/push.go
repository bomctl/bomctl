// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/github/push.go
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
	"fmt"
	"strings"

	"github.com/bomctl/bomctl/internal/pkg/options"
)

func (client *Client) AddFile(pushURL, id string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	if url.Fragment != "" && url.GitRef != "" {
		return client.gitAddFile(pushURL, id, opts)
	}

	return nil
}

func (client *Client) PreparePush(pushURL string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	if url.Fragment != "" && url.GitRef != "" {
		return client.gitPreparePush(pushURL, opts)
	}

	return nil
}

func (client *Client) Push(pushURL string, opts *options.PushOptions) error {
	url := client.Parse(pushURL)
	if url.Fragment != "" && url.GitRef != "" {
		return client.gitPush(pushURL, opts)
	}

	return nil
}

func (client *Client) gitAddFile(pushURL, id string, opts *options.PushOptions) error {
	gitURL := client.assembleGitURL(pushURL)

	err := client.gitClient.AddFile(gitURL, id, opts)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (client *Client) gitPreparePush(pushURL string, opts *options.PushOptions) error {
	gitURL := client.assembleGitURL(pushURL)

	err := client.gitClient.PreparePush(gitURL, opts)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (client *Client) gitPush(pushURL string, opts *options.PushOptions) error {
	gitURL := client.assembleGitURL(pushURL)

	err := client.gitClient.Push(gitURL, opts)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (*Client) assembleGitURL(pushURL string) string {
	urlParts := strings.Split(pushURL, "@")
	gitURL := "git+" + urlParts[0] + ".git@" + urlParts[1]

	return gitURL
}
