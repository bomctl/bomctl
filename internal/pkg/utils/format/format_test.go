package format_test

import (
	"testing"

	"github.com/bomctl/bomctl/internal/pkg/utils/format"
	"github.com/google/go-cmp/cmp"
	"github.com/protobom/protobom/pkg/formats"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		name     string
		fs       string
		encoding string
		want     formats.Format
		wantErr  bool
	}{
		{
			name:     "Parse spdx-2.2 json format",
			fs:       "spdx-2.2",
			encoding: formats.JSON,
			want:     formats.SPDX22JSON,
			wantErr:  false,
		},
		{
			name:     "Parse spdx-2.3 json format",
			fs:       "spdx-2.3",
			encoding: formats.JSON,
			want:     formats.SPDX23JSON,
			wantErr:  false,
		},
		{
			name:     "Parse spdx json format",
			fs:       "spdx",
			encoding: formats.JSON,
			want:     format.DefaultSPDXJSONVersion,
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
