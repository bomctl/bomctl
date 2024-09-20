// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/url/url.go
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

package url

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type (
	ParsedURL struct {
		Scheme   string
		Username string
		Password string
		Hostname string
		Port     string
		GitRef   string
		Path     string
		Query    string
		Fragment string
		Tag      string
		Digest   string
	}

	Parser interface {
		Parse(fetchURL string) *ParsedURL
		RegExp() *regexp.Regexp
	}
)

var ErrParsingURL = errors.New("failed to parse URL")

func (url *ParsedURL) String() string {
	var urlString, pathSep string

	switch url.Scheme {
	case "http", "https", "oci":
		urlString += fmt.Sprintf("%s://", url.Scheme)
		pathSep = "/"
	case "git", "ssh":
		urlString += fmt.Sprintf("%s@", url.Scheme)
		pathSep = ":"
	default:
		pathSep = "/"
	}

	if url.Username != "" && url.Password != "" {
		urlString += fmt.Sprintf("%s:%s@", url.Username, url.Password)
	}

	urlString += url.Hostname

	removeEmpty := func(s string) []string {
		return slices.DeleteFunc(
			[]string{urlString, s}, func(s string) bool { return s == "" },
		)
	}

	urlString = strings.Join(removeEmpty(url.Port), ":")
	urlString = strings.Join(removeEmpty(url.Path), pathSep)
	urlString = strings.Join(removeEmpty(url.GitRef), "@")
	urlString = strings.Join(removeEmpty(url.Query), "?")
	urlString = strings.Join(removeEmpty(url.Fragment), "#")
	urlString = strings.Join(removeEmpty(url.Tag), ":")
	urlString = strings.Join(removeEmpty(url.Digest), "@")

	return urlString
}
