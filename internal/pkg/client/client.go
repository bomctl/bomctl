// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/client/client.go
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

package client

import (
	"errors"
	"fmt"
	neturl "net/url"
	"regexp"
	"strings"

	"github.com/bomctl/bomctl/internal/pkg/client/git"
	"github.com/bomctl/bomctl/internal/pkg/client/github"
	"github.com/bomctl/bomctl/internal/pkg/client/gitlab"
	"github.com/bomctl/bomctl/internal/pkg/client/http"
	"github.com/bomctl/bomctl/internal/pkg/client/oci"
	"github.com/bomctl/bomctl/internal/pkg/netutil"
	"github.com/bomctl/bomctl/internal/pkg/options"
	"github.com/bomctl/bomctl/internal/pkg/sliceutil"
)

var ErrUnsupportedURL = errors.New("failed to parse URL; see `--help` for valid URL patterns")

type (
	Client interface {
		Name() string
	}

	Fetcher interface {
		Client
		Fetch(fetchURL string, opts *options.FetchOptions) ([]byte, error)
		PrepareFetch(url *netutil.URL, auth *netutil.BasicAuth, opts *options.Options) error
	}

	Pusher interface {
		Client
		AddFile(pushURL, id string, opts *options.PushOptions) error
		PreparePush(pushURL string, opts *options.PushOptions) error
		Push(pushURL string, opts *options.PushOptions) error
	}
)

// URL: [scheme:][//[userinfo@]host][/]path[?query][#fragment]
//
// https://username:password@github.com:12345/bomctl/bomctl.git?ref=main#sbom.cdx.json
//
// scheme: https
// userinfo: [username, password]
// host: github.com:12345 (parse port with url.Port())
// path: bomctl/bomctl.git
// fragment: sbom.cdx.json
// Query: ref=main (parse query with url.Query() and get value with query.Get("ref"))
//

func NewFetcher(fetchURL string) (Fetcher, error) {
	var fetch Fetcher

	c, err := New(fetchURL)
	fetch, ok := c.(Fetcher)

	if err != nil || !ok {
		return nil, ErrUnsupportedURL
	}

	return fetch, nil
}

func New(sbomURL string) (Client, error) {
	parsedURL, err := neturl.Parse(sbomURL)
	if err != nil {
		return nil, ErrUnsupportedURL
	}

	return DetermineClient(parsedURL)
}

func DetermineClient(parsedURL *neturl.URL) (Client, error) {
	if client, err := checkScheme(parsedURL); err == nil {
		return client, nil
	}

	gitlabRegex := regexp.MustCompile(`^(www\.)?([a-zA-Z0-9.]+\.)?gitlab\.com$`)
	githubRegex := regexp.MustCompile(`^(www\.)?([a-zA-Z0-9.]+\.)?github\.com$`)

	switch {
	case strings.HasSuffix(parsedURL.Path, ".git") || parsedURL.Fragment != "":
		//nolint:wrapcheck
		return git.Init(parsedURL)
	case githubRegex.MatchString(parsedURL.Host):
		return &github.Client{}, nil
	case gitlabRegex.MatchString(parsedURL.Host):
		return &gitlab.Client{}, nil
	case sliceutil.Any(registrySlice(), func(s string) bool { return strings.Contains(parsedURL.Host, s) }):
		//nolint:wrapcheck
		return oci.Init(parsedURL)
	case parsedURL.Scheme == "https" || parsedURL.Scheme == "http":
		return &http.Client{}, nil
	default:
		return nil, fmt.Errorf("%w", errors.ErrUnsupported)
	}
}

func checkScheme(parsedURL *neturl.URL) (Client, error) {
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// check for explicit selection via scheme
	switch parsedURL.Scheme {
	case "git":
		return &git.Client{}, nil
	case "github":
		return &github.Client{}, nil
	case "gitlab":
		return &gitlab.Client{}, nil
	case "oci", "oci-archive", "docker", "docker-archive":
		//nolint:wrapcheck
		return oci.Init(parsedURL)
	default:
		return nil, fmt.Errorf("%w", errors.ErrUnsupported)
	}
}

func registrySlice() []string {
	// maybe check for env variable like BOMCTL_REGISTRY or something and append to slice if it exists
	return []string{"docker.io", "gcr.io", "ghcr.io", "quay.io"}
}
