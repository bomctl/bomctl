package format_test

import (
	"strings"
	"testing"

	"github.com/bom-squad/protobom/pkg/formats"
	"github.com/bomctl/bomctl/internal/pkg/utils/format"
	"github.com/google/go-cmp/cmp"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		name     string
		fs       string
		encoding string
		want     *format.Format
		wantErr  bool
	}{
		{
			name:     "Parse spdx-2.2 json format",
			fs:       "spdx-2.2",
			encoding: formats.JSON,
			want:     &format.Format{formats.SPDX22JSON},
			wantErr:  false,
		},
		{
			name:     "Parse spdx-2.3 json format",
			fs:       "spdx-2.3",
			encoding: formats.JSON,
			want:     &format.Format{formats.SPDX23JSON},
			wantErr:  false,
		},
		{
			name:     "Parse spdx json format",
			fs:       "spdx",
			encoding: formats.JSON,
			want:     &format.Format{format.DefaultSPDXJSONVersion},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := format.Parse(tt.fs, tt.encoding)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Inverse(t *testing.T) {
	tests := []struct {
		name    string
		format  *format.Format
		want    *format.Format
		wantErr bool
	}{
		{
			name:    "Inverse of spdx json format",
			format:  &format.Format{formats.SPDX22JSON},
			want:    &format.Format{format.DefaultCycloneDXVersion},
			wantErr: false,
		},
		{
			name:    "Inverse of cyclonedx json format",
			format:  &format.Format{formats.CDX15JSON},
			want:    &format.Format{format.DefaultSPDXJSONVersion},
			wantErr: false,
		},
		{
			name:    "Inverse of spdx tv format",
			format:  &format.Format{formats.SPDX22TV},
			want:    &format.Format{format.DefaultCycloneDXVersion},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.format.Inverse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Inverse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Inverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Detect(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *format.Format
		wantErr bool
	}{
		{
			name:    "Detect spdx-2.2 json format",
			input:   `{"SPDXVersion": "SPDX-2.2", "DataLicense": "CC0-1.0", "SPDXID": "SPDXRef-DOCUMENT", "DocumentName": "SPDX-2.2", "DocumentNamespace": "http://spdx.org/spdxdocs/spdx-v2.2-3c4714e6-a7b1-4574-abb8-861149cbc590", "Creator": "Tool: SPDX-Java-Tools v0.2.5", "Created": "2020-07-23T18:30:22Z"}`, //nolint:lll
			want:    &format.Format{formats.SPDX22JSON},
			wantErr: false,
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := format.Detect(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}
