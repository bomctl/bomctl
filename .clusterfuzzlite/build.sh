#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
# SPDX-FileName: .clusterfuzzlite/build.sh
# SPDX-FileType: SOURCE
# SPDX-License-Identifier: Apache-2.0
# -----------------------------------------------------------------------------
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# -----------------------------------------------------------------------------

set -euo pipefail

# Prepare executable and dependencies
go mod tidy
go get github.com/AdamKorcz/go-118-fuzz-build/testing

compile_native_go_fuzzer github.com/bomctl/bomctl/internal/pkg/fetch_test FuzzFetch FuzzFetch
compile_native_go_fuzzer github.com/bomctl/bomctl/internal/pkg/export_test FuzzExport FuzzExport
compile_native_go_fuzzer github.com/bomctl/bomctl/internal/pkg/merge_test FuzzMerge FuzzMerge
compile_native_go_fuzzer github.com/bomctl/bomctl/internal/pkg/push_test FuzzPush FuzzPush
