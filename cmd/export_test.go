// -----------------------------------------------------------------------------
// SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
// SPDX-FileName: cmd/export_test.go
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

package cmd_test

import (
	"strings"
	"testing"

	"github.com/protobom/protobom/pkg/formats"
	"github.com/stretchr/testify/require"

	"github.com/bomctl/bomctl/cmd"
)

func Test_ParseFormat(t *testing.T) {
	t.Parallel()

	availableEncodings := cmd.EncodingOptions()
	availableFormats := cmd.FormatOptions()

	for _, format := range availableFormats {
		formatBase := getFormatBase(format)
		for _, encoding := range availableEncodings[formatBase] {
			_, err := cmd.ParseFormat(format, encoding)
			if err != nil {
				t.Logf("unexpected error while parsing format: %s with encoding: %s. Err: %v", format, encoding, err)
				t.FailNow()
			}
		}
	}
}

func getFormatBase(fmt string) string {
	if strings.Contains(fmt, formats.CDXFORMAT) {
		return formats.CDXFORMAT
	}

	return formats.SPDXFORMAT
}

func Test_ParseSpecificFormat(t *testing.T) {
	t.Parallel()

	for _, data := range []struct {
		name      string
		format    string
		encoding  string
		expected  formats.Format
		shouldErr bool
	}{
		{
			name:      "Default spdx",
			format:    formats.SPDXFORMAT,
			encoding:  formats.JSON,
			expected:  formats.SPDX23JSON,
			shouldErr: false,
		},
		{
			name:      "spdx 2.3",
			format:    "spdx-2.3",
			encoding:  formats.JSON,
			expected:  formats.SPDX23JSON,
			shouldErr: false,
		},
		{
			name:      "xml spdx",
			format:    formats.SPDXFORMAT,
			encoding:  formats.XML,
			expected:  formats.EmptyFormat,
			shouldErr: true,
		},
		{
			name:      "non-supported spdx",
			format:    "spdx-0.7",
			encoding:  formats.XML,
			expected:  formats.EmptyFormat,
			shouldErr: true,
		},
		{
			name:      "Default CDX",
			format:    formats.CDXFORMAT,
			encoding:  formats.JSON,
			expected:  formats.CDX15JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.0 json",
			format:    "cyclonedx-1.0",
			encoding:  formats.JSON,
			expected:  formats.CDX10JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.1 json",
			format:    "cyclonedx-1.1",
			encoding:  formats.JSON,
			expected:  formats.CDX11JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.2 json",
			format:    "cyclonedx-1.2",
			encoding:  formats.JSON,
			expected:  formats.CDX12JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.3 json",
			format:    "cyclonedx-1.3",
			encoding:  formats.JSON,
			expected:  formats.CDX13JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.4 json",
			format:    "cyclonedx-1.4",
			encoding:  formats.JSON,
			expected:  formats.CDX14JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.5 json",
			format:    "cyclonedx-1.5",
			encoding:  formats.JSON,
			expected:  formats.CDX15JSON,
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.0 xml",
			format:    "cyclonedx-1.0",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.0"),
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.1 xml",
			format:    "cyclonedx-1.1",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.1"),
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.2 xml",
			format:    "cyclonedx-1.2",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.2"),
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.3 xml",
			format:    "cyclonedx-1.3",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.3"),
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.4 xml",
			format:    "cyclonedx-1.4",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.4"),
			shouldErr: false,
		},
		{
			name:      "cyclonedx 1.5 xml",
			format:    "cyclonedx-1.5",
			encoding:  formats.XML,
			expected:  formats.Format("application/vnd.cyclonedx+xml;version=1.5"),
			shouldErr: false,
		},
		{
			name:      "unknown encoding cyclonedx",
			format:    formats.CDXFORMAT,
			encoding:  "other",
			expected:  formats.EmptyFormat,
			shouldErr: true,
		},
		{
			name:      "non-supported cyclonedx xml",
			format:    "cyclonedx-7",
			encoding:  formats.XML,
			expected:  formats.EmptyFormat,
			shouldErr: true,
		},
		{
			name:      "non-supported cyclonedx json",
			format:    "cyclonedx-1.23",
			encoding:  formats.JSON,
			expected:  formats.EmptyFormat,
			shouldErr: true,
		},
	} {
		got, err := cmd.ParseFormat(data.format, data.encoding)
		if !data.shouldErr {
			require.NoError(t, err)
		}

		require.Equal(t, data.expected, got)
	}
}
