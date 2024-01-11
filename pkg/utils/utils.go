// -------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
// SPDX-FileName: pkg/utils/utils.go
// SPDX-FileType: SOURCE
// SPDX-License-Identifier: Apache-2.0
// -------------------------------------------------------
package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

var urlPattern = regexp.MustCompile(`^(` +
	`(?P<scheme>git|http(s)?|oci|ssh)(?:@|://))?(//)?((` +
	`(?P<username>[^:]+)(?::(?P<password>[^@]+)?)?@)?` +
	`(?P<hostname>[^@/?#:]*)(?::(?P<port>\d+)?)?)?` +
	`(/?(?P<path>[^?#]*))(\?(?P<query>[^#]*))?(#(?P<fragment>.*))?`)

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
	Path     string
	Query    string
	Fragment string
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

func ParseURL(url string) parsedURL {
	matches := urlPattern.FindStringSubmatch(url)

	return parsedURL{
		Scheme:   matches[urlPattern.SubexpIndex("scheme")],
		Username: matches[urlPattern.SubexpIndex("username")],
		Password: matches[urlPattern.SubexpIndex("password")],
		Hostname: matches[urlPattern.SubexpIndex("hostname")],
		Port:     matches[urlPattern.SubexpIndex("port")],
		Path:     matches[urlPattern.SubexpIndex("path")],
		Query:    matches[urlPattern.SubexpIndex("query")],
		Fragment: matches[urlPattern.SubexpIndex("fragment")],
	}
}
