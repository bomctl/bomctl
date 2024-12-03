// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: internal/pkg/merge/merge_fuzz.go
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

package merge_test

import (
	"context"
	"testing"

	"github.com/bomctl/bomctl/internal/pkg/merge"
	"github.com/bomctl/bomctl/internal/pkg/options"
)

const (
	id1 = "urn:uuid:f360ad8b-dc41-4256-afed-337a04dff5db"
	id2 = "urn:uuid:f360ad8b-dc41-4256-afed-337a04dff5dc"
)

func FuzzMerge(f *testing.F) {
	f.Add(id1, id2)
	f.Fuzz(func(t *testing.T, id1, id2 string) {
		_, err := merge.Merge(
			[]string{id1, id2},
			&options.MergeOptions{Options: options.New().WithContext(context.Background())},
		)

		if err == nil {
			t.Errorf("%s", err)
		}
	})
}
