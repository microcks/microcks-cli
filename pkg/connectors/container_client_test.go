package connectors

import (
	"errors"
	"io"
	"strings"
	"testing"
)

type failingReader struct{}

func (f failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestCopyImagePullOutputReturnsReadError(t *testing.T) {
	err := copyImagePullOutput(failingReader{})

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to read image pull output") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCopyImagePullOutputSucceeds(t *testing.T) {
	err := copyImagePullOutput(strings.NewReader(""))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

var _ io.Reader = failingReader{}
