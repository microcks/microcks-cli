package rand

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{name: "zero length", length: 0},
		{name: "single character", length: 1},
		{name: "short string", length: 8},
		{name: "longer string", length: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := String(tt.length)
			if err != nil {
				t.Fatalf("String(%d) returned error: %v", tt.length, err)
			}
			if len(got) != tt.length {
				t.Fatalf("String(%d) length = %d, want %d", tt.length, len(got), tt.length)
			}
			for i := 0; i < len(got); i++ {
				if !strings.Contains(letterBytes, string(got[i])) {
					t.Fatalf("String(%d) contains invalid character %q at index %d", tt.length, got[i], i)
				}
			}
		})
	}
}

func TestStringProducesDifferentValues(t *testing.T) {
	t.Parallel()

	const length = 32

	first, err := String(length)
	if err != nil {
		t.Fatalf("first String(%d) returned error: %v", length, err)
	}

	second, err := String(length)
	if err != nil {
		t.Fatalf("second String(%d) returned error: %v", length, err)
	}

	if first == second {
		t.Fatalf("expected two generated strings to differ, both were %q", first)
	}
}

func TestStringFromCharset(t *testing.T) {
	t.Parallel()

	const charset = "01"

	got, err := StringFromCharset(16, charset)
	if err != nil {
		t.Fatalf("StringFromCharset returned error: %v", err)
	}
	if len(got) != 16 {
		t.Fatalf("StringFromCharset length = %d, want 16", len(got))
	}
	for i := 0; i < len(got); i++ {
		if !strings.Contains(charset, string(got[i])) {
			t.Fatalf("StringFromCharset contains invalid character %q at index %d", got[i], i)
		}
	}
}

func TestStringFromCharsetZeroLength(t *testing.T) {
	t.Parallel()

	got, err := StringFromCharset(0, "abc")
	if err != nil {
		t.Fatalf("StringFromCharset(0) returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("StringFromCharset(0) = %q, want empty string", got)
	}
}

func TestStringFromCharsetSingleCharacter(t *testing.T) {
	t.Parallel()

	got, err := StringFromCharset(10, "x")
	if err != nil {
		t.Fatalf("StringFromCharset returned error: %v", err)
	}
	if got != "xxxxxxxxxx" {
		t.Fatalf("StringFromCharset with single-character charset = %q, want %q", got, "xxxxxxxxxx")
	}
}

func TestStringFromCharsetEmptyCharset(t *testing.T) {
	t.Parallel()

	got, err := StringFromCharset(5, "")
	if err == nil {
		t.Fatal("StringFromCharset with empty charset expected error, got nil")
	}
	if got != "" {
		t.Fatalf("StringFromCharset with empty charset = %q, want empty string on error", got)
	}
}
