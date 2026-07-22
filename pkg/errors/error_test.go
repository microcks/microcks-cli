package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestWrapNilReturnsNil(t *testing.T) {
	if Wrap(KindAPI, nil) != nil {
		t.Fatal("Wrap of a nil error should return nil")
	}
}

func TestKindOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"nil", nil, KindGeneric},
		{"plain error", stderrors.New("boom"), KindGeneric},
		{"wrapped", Wrap(KindConnection, stderrors.New("refused")), KindConnection},
		{"wrapped again", fmt.Errorf("outer: %w", Wrap(KindNotFound, stderrors.New("404"))), KindNotFound},
	}
	for _, c := range cases {
		if got := KindOf(c.err); got != c.want {
			t.Errorf("%s: KindOf = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestWrapfPreservesKindAndMessage(t *testing.T) {
	err := Wrapf(KindEnvironment, "docker unreachable on attempt %d", 2)
	if KindOf(err) != KindEnvironment {
		t.Errorf("KindOf = %v, want KindEnvironment", KindOf(err))
	}
	if got := err.Error(); got != "docker unreachable on attempt 2" {
		t.Errorf("Error() = %q", got)
	}
}
