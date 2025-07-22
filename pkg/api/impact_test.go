package api

import (
	"testing"
)

func TestParseComponentImpact(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ComponentImpact
		expectErr bool
	}{
		{
			name:      "valid full impact name",
			input:     "operational api",
			want:      ComponentImpact{Component: "api", Impact: ImpactTypeOperational},
			expectErr: false,
		},
		{
			name:      "valid shorthand impact",
			input:     "op web",
			want:      ComponentImpact{Component: "web", Impact: ImpactTypeOperational},
			expectErr: false,
		},
		{
			name:      "another valid shorthand impact",
			input:     "u db",
			want:      ComponentImpact{Component: "db", Impact: ImpactTypeUnderMaintenance},
			expectErr: false,
		},
		{
			name:      "invalid impact type",
			input:     "unknown impact",
			expectErr: true,
		},
		{
			name:      "invalid format with only impact",
			input:     "operational",
			expectErr: true,
		},
		{
			name:      "invalid format with only component",
			input:     "api",
			expectErr: true,
		},
		{
			name:      "empty input",
			input:     "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComponentImpact(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseComponentImpact() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && got != tt.want {
				t.Errorf("ParseComponentImpact() got = %v, want %v", got, tt.want)
			}
		})
	}
}
