// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/fuzz_test.go
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

package fetch

import (
	"testing"

	"github.com/bomctl/bomctl/internal/pkg/fetch"
)

func FuzzFetch(f *testing.F) {
	f.Add([]byte("https://raw.githubusercontent.com/bomctl/bomctl-playground/main/examples/bomctl-container-image/bomctl_bomctl_v0.3.0.cdx.json"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, err := fetch.Fetch(string(data), nil)
		if err == nil {
			t.Errorf("ERROR:")
		}
	})
}
