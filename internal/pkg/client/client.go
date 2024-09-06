// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/client.go
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
package client

import (
	"errors"
	"fmt"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/client/http"
	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/url"
)

var errUnsupportedURL = errors.New("failed to parse URL; see `--help` for valid URL patterns")

type Client interface {
	url.Parser
	AddFile(string, string, *options.PushOptions) error
	Name() string
	Fetch(string, *options.FetchOptions) ([]byte, error)
	PreparePush(string, *options.PushOptions) error
	Push(string, string, *options.PushOptions) error
}

func New(sbomURL string) (Client, error) {
	for _, client := range []Client{&git.Client{}, &http.Client{}, &oci.Client{}} {
		if parsedURL := client.Parse(sbomURL); parsedURL != nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", errUnsupportedURL, sbomURL)
}
