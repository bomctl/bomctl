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
)

type basicAuthCredentials struct {
	username string
	password string
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
