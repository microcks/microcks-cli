package rand

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString_ValidLengths(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"zero length", 0},
		{"single char", 1},
		{"medium length", 24},
		{"PKCE-style length", 43},
		{"long", 128},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String(tt.n)
			require.NoError(t, err)
			assert.Len(t, got, tt.n)
			for i := 0; i < len(got); i++ {
				assert.True(t, strings.IndexByte(letterBytes, got[i]) >= 0,
					"byte %q at position %d not in default letter charset", got[i], i)
			}
		})
	}
}

func TestString_NegativeLengthReturnsError(t *testing.T) {
	got, err := String(-1)
	require.Error(t, err)
	assert.Equal(t, "", got)
}

func TestStringFromCharset_ValidInputs(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		charset string
	}{
		{"empty length, empty charset", 0, ""},
		{"empty length, non-empty charset", 0, "abc"},
		{"single byte from single-byte charset", 1, "x"},
		{"PKCE verifier charset", 43, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringFromCharset(tt.n, tt.charset)
			require.NoError(t, err)
			assert.Len(t, got, tt.n)
			for i := 0; i < len(got); i++ {
				assert.True(t, strings.IndexByte(tt.charset, got[i]) >= 0,
					"byte %q at position %d not in charset %q", got[i], i, tt.charset)
			}
		})
	}
}

func TestStringFromCharset_InvalidInputsReturnError(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		charset string
	}{
		{"negative length, non-empty charset", -1, "abc"},
		{"negative length, empty charset", -1, ""},
		{"positive length, empty charset", 1, ""},
		{"large positive length, empty charset", 100, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringFromCharset(tt.n, tt.charset)
			require.Error(t, err)
			assert.Equal(t, "", got)
		})
	}
}
