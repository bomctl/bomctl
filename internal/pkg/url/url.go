// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: internal/pkg/url/url.go
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
package url

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type BasicAuth struct {
	Username, Password string
}

func (auth *BasicAuth) Encode() string {
	if auth == nil {
		return ""
	}

	data := []byte(strings.Join([]string{auth.Username, auth.Password}, ":"))

	return base64.URLEncoding.EncodeToString(data)
}

func (*BasicAuth) Name() string {
	return "http-basic-auth"
}

func (auth *BasicAuth) SetAuth(request *http.Request) {
	if auth == nil {
		return
	}

	request.SetBasicAuth(auth.Username, auth.Password)
}

func (auth *BasicAuth) String() string {
	masked := "*******"
	if auth.Password == "" {
		masked = "<empty>"
	}

	return fmt.Sprintf("Authorization: Basic %s:%s", auth.Username, masked)
}

type ParsedURL struct {
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

func (url *ParsedURL) String() string { //nolint:cyclop
	var (
		urlBytes []byte
		pathSep  string
	)

	switch url.Scheme {
	case "http", "https", "oci":
		urlBytes = append(urlBytes, fmt.Sprintf("%s://", url.Scheme)...)
		pathSep = "/"
	case "git", "ssh":
		urlBytes = append(urlBytes, fmt.Sprintf("%s@", url.Scheme)...)
		pathSep = ":"
	}

	if url.Username != "" && url.Password != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("%s:%s@", url.Username, url.Password)...)
	}

	urlBytes = append(urlBytes, url.Hostname...)

	if url.Path != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("%s%s", pathSep, url.Path)...)
	}

	if url.GitRef != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("@%s", url.GitRef)...)
	}

	if url.Query != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("?%s", url.Query)...)
	}

	if url.Fragment != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("#%s", url.Fragment)...)
	}

	if url.Tag != "" {
		urlBytes = append(urlBytes, fmt.Sprintf(":%s", url.Tag)...)
	}

	if url.Digest != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("@%s", url.Digest)...)
	}

	return string(urlBytes)
}

type Parser interface {
	Parse(fetchURL string) *ParsedURL
	RegExp() *regexp.Regexp
}
