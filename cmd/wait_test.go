package cmd

import "testing"

func TestParseWaitForMilliseconds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "milli", input: "500milli", want: 500},
		{name: "sec", input: "5sec", want: 5000},
		{name: "min", input: "2min", want: 120000},
		{name: "invalid unit", input: "5hours", wantErr: true},
		{name: "invalid number", input: "xsec", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWaitForMilliseconds(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseWaitForMilliseconds(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWaitForMilliseconds(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseWaitForMilliseconds(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
