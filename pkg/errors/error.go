package errors

import (
	"log"
	"os"
)

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

func CheckError(err error) {
	if err != nil {
		Fatal(ErrorGeneric, err)
	}
}

func CheckConfigNil(isNil bool, path string) {
	if isNil {
		Fatal(ErrorGeneric, "No contexts defined in "+path)
	}
}


// Fatal is a wrapper for log.Fatal() to exit with custom code
func Fatal(exitcode int, args ...interface{}) {
	log.Println(args...)
	os.Exit(exitcode)
}


