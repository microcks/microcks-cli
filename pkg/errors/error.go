package errors

import (
	stderrors "errors"
	"fmt"
)

// Kind classifies why an operation failed. The library returns kinds; the cmd
// layer maps them to exit codes, so pkg/* never depends on exit codes and stays
// safe to embed.
type Kind int

const (
	// KindGeneric is an unclassified failure. It is the zero value, so an
	// unwrapped error is treated as generic.
	KindGeneric Kind = iota
	// KindUsage is a bad argument or flag supplied by the user.
	KindUsage
	// KindConnection is a failure to reach the Microcks or Keycloak endpoint.
	KindConnection
	// KindAPI is a server that rejected the request or returned an unusable response.
	KindAPI
	// KindNotFound is a requested remote resource that does not exist.
	KindNotFound
	// KindEnvironment is a local precondition not met — container runtime down,
	// image unpullable, or ephemeral server not ready. Not KindConnection, which
	// is about reaching the Microcks server.
	KindEnvironment
)

// KindError wraps an error with a Failure Kind.
type KindError struct {
	Kind Kind
	Err  error
}

func (e *KindError) Error() string { return e.Err.Error() }
func (e *KindError) Unwrap() error { return e.Err }

// Wrap tags err with a Failure Kind. It returns nil when err is nil, so it is
// safe to write `return errors.Wrap(KindAPI, doThing())`.
func Wrap(kind Kind, err error) error {
	if err == nil {
		return nil
	}
	return &KindError{Kind: kind, Err: err}
}

// Wrapf builds a kind-tagged error from a format string.
func Wrapf(kind Kind, format string, a ...any) error {
	return &KindError{Kind: kind, Err: fmt.Errorf(format, a...)}
}

// KindOf reports the Failure Kind carried by err's chain, defaulting to
// KindGeneric when none is present (including when err is nil).
func KindOf(err error) Kind {
	var ke *KindError
	if stderrors.As(err, &ke) {
		return ke.Kind
	}
	return KindGeneric
}

// ErrTestFailed signals a completed test run whose result does not conform to the
// contract. It is not a failure to *run*: the command has already rendered the
// result, so the CLI exits non-zero without printing this sentinel. See cmd.Handle.
var ErrTestFailed = stderrors.New("contract test failed")
