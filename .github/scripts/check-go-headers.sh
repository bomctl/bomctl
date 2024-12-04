#!/usr/bin/env bash

set -euo pipefail

# Find all Golang files in project.
go_files=$(git ls-files '*.go')

# File header pattern lines (contains format specifier for SPDX-FileName value)
header_template='// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © * bomctl a Series of LF Projects, LLC
// SPDX-FileName: %s
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

package *'

fix_files=()

for file in $go_files; do
  # Expand format specifiers in header template.
  # shellcheck disable=SC2059
  printf -v header "$header_template" "$file"

  # Strip go build tags
  content=$(sed '/\/\/go:build/,+2d' < "$file")

  # shellcheck disable=SC2053
  if [[ $(echo "$content" | head --lines=20) != $header ]]; then
    fix_files+=("$file")
  fi
done

if [[ ${#fix_files[@]} -gt 0 ]]; then
  echo "The following files have SPDX license header issues:"
  printf "\t%s\n" "${fix_files[@]}"
  exit 1
fi
