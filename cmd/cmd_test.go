package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestStartCommandPortValidation(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{name: "lowest valid port", port: "1", wantErr: false},
		{name: "default port", port: "8585", wantErr: false},
		{name: "highest valid port", port: "65535", wantErr: false},
		{name: "non numeric port", port: "abc", wantErr: true},
		{name: "zero port", port: "0", wantErr: true},
		{name: "negative port", port: "-1", wantErr: true},
		{name: "port above range", port: "65536", wantErr: true},
		{name: "empty port", port: "", wantErr: true},
		{name: "port with spaces", port: " 8080", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStartCommand(nil)
			cmd.SetArgs([]string{"--port", tt.port})

			executed := false
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				executed = true
				return nil
			}

			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				require.False(t, executed, "RunE should not execute when validation fails")
			} else {
				require.NoError(t, err)
				require.True(t, executed, "RunE should execute when validation passes")
			}
		})
	}
}

func TestLoginCommandSsoPortValidation(t *testing.T) {
	tests := []struct {
		name    string
		ssoPort int
		wantErr bool
	}{
		{name: "lowest valid port", ssoPort: 1, wantErr: false},
		{name: "default port", ssoPort: 58085, wantErr: false},
		{name: "highest valid port", ssoPort: 65535, wantErr: false},
		{name: "zero port", ssoPort: 0, wantErr: true},
		{name: "negative port", ssoPort: -1, wantErr: true},
		{name: "port above range", ssoPort: 65536, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLoginCommand(nil)
			cmd.SetArgs([]string{"--sso-port", fmt.Sprintf("%d", tt.ssoPort), "http://localhost:8080"})

			executed := false
			originalRun := cmd.Run
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				executed = true
				return nil
			}

			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				require.False(t, executed, "RunE should not execute when validation fails")
			} else {
				require.NoError(t, err)
				require.True(t, executed, "RunE should execute when validation passes")
			}
			_ = originalRun
		})
	}
}
