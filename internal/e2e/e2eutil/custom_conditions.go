// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/e2e/e2eutil/custom_conditions.go
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

package e2eutil

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	versions "github.com/hashicorp/go-version"

	"github.com/bomctl/bomctl/cmd"
)

const (
	minVersionArgs = 2
	minExistsArgs  = 3
)

var (
	errVersionParse     = errors.New("failed to parse version")
	errUnknownCondition = errors.New("unrecognized condition")
	errArgument         = errors.New("missing required argument")
	errIntConversion    = errors.New("failed to convert to Int")
	errRequiredVersion  = errors.New("current GO version insufficient")
	errFileNotFound     = errors.New("file not found within given time")
)

// customConditions is a testscript function that handles all the conditions defined for this test.
func CustomConditions(condition string) (bool, error) {
	// assumes arguments are separated by a colon (":")
	elements := strings.Split(condition, ":")
	if len(elements) == 0 {
		return false, errArgument
	}

	name := elements[0]
	switch name {
	case "version_is_at_least":
		return versionIsAtLeast(elements)
	case "exists_within_seconds":
		return existsWithinSeconds(elements)
	default:
		return false, errUnknownCondition
	}
}

func versionGreaterOrEqual(version1, version2 string) (bool, error) {
	ver1, err := versions.NewVersion(version1)
	if err != nil {
		return false, errVersionParse
	}

	ver2, err := versions.NewVersion(version2)
	if err != nil {
		return false, errVersionParse
	}

	return ver1.GreaterThanOrEqual(ver2), nil
}

func versionIsAtLeast(elements []string) (bool, error) {
	if len(elements) < minVersionArgs {
		return false, errArgument
	}

	version := elements[1]

	result, err := versionGreaterOrEqual(cmd.Version, version)
	if result && os.Getenv("WCDEBUG") != "" {
		err = errRequiredVersion
	}

	return result, err
}

func existsWithinSeconds(elements []string) (bool, error) {
	if len(elements) < minExistsArgs {
		return false, errArgument
	}

	fileName := elements[1]

	delay, err := strconv.Atoi(elements[2])
	if err != nil {
		return false, errIntConversion
	}

	if delay == 0 {
		return fileExists(fileName), nil
	}

	elapsed := 0
	for elapsed < delay {
		time.Sleep(time.Second)

		if fileExists(fileName) {
			return true, nil
		}

		elapsed++
	}

	return false, errFileNotFound
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	return !os.IsNotExist(err)
}
