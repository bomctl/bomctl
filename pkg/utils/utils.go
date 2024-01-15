// ------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: pkg/utils/utils.go
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
package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

var urlPattern = regexp.MustCompile(
	fmt.Sprintf("%s%s%s%s%s",
		`^//|(?P<scheme>git|http(s)?|oci|ssh)(@|(\+http(s)?)?://)`,
		`(((?P<username>[^:]+)(?::(?P<password>[^@]+))?:@)?`,
		`(?P<hostname>[^@/?#:]*)(?::(?P<port>\d+)?)?)?`,
		`(/?(?P<path>[^@?#]*))(?:@(?P<gitRef>[^#]+))?`,
		`(\?(?P<query>[^#]*))?(#(?P<fragment>.*))?`,
	),
)

type basicAuthCredentials struct {
	username string
	password string
}

type parsedURL struct {
	Scheme   string
	Username string
	Password string
	Hostname string
	Port     string
	GitRef   string
	Path     string
	Query    string
	Fragment string
}

func (url *parsedURL) String() string {
	var urlBytes []byte
	pathSep := ""

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

	if url.Query != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("?%s", url.Query)...)
	}

	if url.Fragment != "" {
		urlBytes = append(urlBytes, fmt.Sprintf("#%s", url.Fragment)...)
	}

	return string(urlBytes)
}

func basicAuth(username, password string) string {
	data := []byte(username + ":" + password)
	return base64.URLEncoding.EncodeToString(data)
}

func DownloadHTTP(url, filepath string, auth *basicAuthCredentials) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer out.Close()

	client := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if auth != nil {
		req.Header.Add("Authorization", "Basic "+basicAuth(auth.username, auth.password))
	}

	// Get the data
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	defer resp.Body.Close()

	// Write the response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func ParseURL(url string) *parsedURL {
	matches := urlPattern.FindStringSubmatch(url)

	return &parsedURL{
		Scheme:   matches[urlPattern.SubexpIndex("scheme")],
		Username: matches[urlPattern.SubexpIndex("username")],
		Password: matches[urlPattern.SubexpIndex("password")],
		Hostname: matches[urlPattern.SubexpIndex("hostname")],
		Port:     matches[urlPattern.SubexpIndex("port")],
		GitRef:   matches[urlPattern.SubexpIndex("gitRef")],
		Path:     matches[urlPattern.SubexpIndex("path")],
		Query:    matches[urlPattern.SubexpIndex("query")],
		Fragment: matches[urlPattern.SubexpIndex("fragment")],
	}
}
