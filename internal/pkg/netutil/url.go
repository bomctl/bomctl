// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/netutil/url.go
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

package netutil

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type (
	Parser interface {
		Parse(url string) *URL
		RegExp() *regexp.Regexp
	}

	URL struct {
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
)

var ErrParsingURL = errors.New("failed to parse URL")

func (url *URL) String() string {
	var urlString, pathSep string

	switch url.Scheme {
	case "http", "https", "oci":
		urlString += url.Scheme + "://"
		pathSep = "/"
	case "git", "ssh":
		urlString += url.Scheme + "@"
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
