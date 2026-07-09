package errors

import (
	stderrors "errors"
	"fmt"
	"log"
	"os"
)

// Deprecated: these numeric codes and the Check*/Fatal helpers below are the
// legacy exit mechanism. New code classifies failures with a Kind (see Wrap) and
// lets cmd.Handle map Kind -> exit code. Kept as a shim until every call site is
// migrated, then removed.
const (
	// ErrorCommandSpecific is reserved for command specific indications
	ErrorCommandSpecific = 1
	// ErrorConnectionFailure is returned on connection failure to API endpoint
	ErrorConnectionFailure = 11
	// ErrorAPIResponse is returned on unexpected API response, i.e. authorization failure
	ErrorAPIResponse = 12
	// ErrorResourceDoesNotExist is returned when the requested resource does not exist
	ErrorResourceDoesNotExist = 13
	// ErrorGeneric is returned for generic error
	ErrorGeneric = 20
)

// Deprecated: return errors.Wrap(kind, err) from a RunE command instead.
func CheckError(err error) {
	if err != nil {
		Fatal(ErrorGeneric, err)
	}
}

// Deprecated: return a KindNotFound-wrapped error instead.
func CheckConfigNil(isNil bool, path string) {
	if isNil {
		Fatal(ErrorGeneric, "No contexts defined in "+path)
	}
}

// Deprecated: only main/cmd.Handle should exit the process. Fatal is a wrapper
// for log.Fatal() to exit with a custom code.
func Fatal(exitcode int, args ...interface{}) {
	log.Println(args...)
	os.Exit(exitcode)
}

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
