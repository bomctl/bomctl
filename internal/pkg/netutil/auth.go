// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/netutil/auth.go
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
	"fmt"
	"os"
	"path/filepath"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jdx/go-netrc"
)

type BasicAuth struct {
	githttp.BasicAuth
}

func (auth *BasicAuth) UseNetRC(hostname string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	authFile, err := netrc.Parse(filepath.Join(home, ".netrc"))
	if err != nil {
		return fmt.Errorf("failed to parse .netrc file: %w", err)
	}

	// Use credentials in .netrc if entry for the hostname is found
	if machine := authFile.Machine(hostname); machine != nil {
		auth.Username = machine.Get("login")
		auth.Password = machine.Get("password")
	}

	return nil
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{githttp.BasicAuth{Username: username, Password: password}}
}
