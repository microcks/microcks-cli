package cmd

import (
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/stretchr/testify/assert"
)

func TestParseWaitFor(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      int64
		expectedError string
	}{
		{
			name:     "Valid milliseconds",
			input:    "500milli",
			expected: 500,
		},
		{
			name:     "Valid seconds",
			input:    "5sec",
			expected: 5000,
		},
		{
			name:     "Valid minutes",
			input:    "2min",
			expected: 120000,
		},
		{
			name:          "Invalid format",
			input:         "5hours",
			expectedError: "--waitFor format is wrong. Accepted units are: milli, sec, min (e.g. 500milli, 30sec, 5min)",
		},
		{
			name:          "Invalid number milli",
			input:         "abcmilli",
			expectedError: "--waitFor value \"abcmilli\" is not a valid number",
		},
		{
			name:          "Invalid number sec",
			input:         "xyzsec",
			expectedError: "--waitFor value \"xyzsec\" is not a valid number",
		},
		{
			name:          "Invalid number min",
			input:         "1.5min",
			expectedError: "--waitFor value \"1.5min\" is not a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseWaitFor(tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNewTestCommand(t *testing.T) {
	clientOpts := &connectors.ClientOptions{}
	cmd := NewTestCommand(clientOpts)

	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Run tests on Microcks", cmd.Short)

	waitForFlag := cmd.Flags().Lookup("waitFor")
	assert.NotNil(t, waitForFlag)
	assert.Equal(t, "5sec", waitForFlag.DefValue)

	secretNameFlag := cmd.Flags().Lookup("secretName")
	assert.NotNil(t, secretNameFlag)

	filteredOperationsFlag := cmd.Flags().Lookup("filteredOperations")
	assert.NotNil(t, filteredOperationsFlag)

	operationsHeadersFlag := cmd.Flags().Lookup("operationsHeaders")
	assert.NotNil(t, operationsHeadersFlag)

	oAuth2ContextFlag := cmd.Flags().Lookup("oAuth2Context")
	assert.NotNil(t, oAuth2ContextFlag)
}
