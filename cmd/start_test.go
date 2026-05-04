package cmd

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateStartPort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{name: "lowest valid port", port: "1"},
		{name: "default port", port: "8585"},
		{name: "highest valid port", port: "65535"},
		{name: "non numeric port", port: "abc", wantErr: true},
		{name: "zero port", port: "0", wantErr: true},
		{name: "negative port", port: "-1", wantErr: true},
		{name: "port above range", port: "65536", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStartPort(tt.port)
			if tt.wantErr {
				require.EqualError(t, err, "--port must be a number between 1 and 65535, got "+strconv.Quote(tt.port))
				return
			}
			require.NoError(t, err)
		})
	}
}
