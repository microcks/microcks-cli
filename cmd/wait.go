package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

func parseWaitForMilliseconds(waitFor string) (int64, error) {
	if strings.HasSuffix(waitFor, "milli") {
		n, err := strconv.ParseInt(waitFor[:len(waitFor)-5], 0, 64)
		if err != nil {
			return 0, fmt.Errorf("--waitFor value %q is not a valid number", waitFor)
		}
		return n, nil
	}
	if strings.HasSuffix(waitFor, "sec") {
		n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
		if err != nil {
			return 0, fmt.Errorf("--waitFor value %q is not a valid number", waitFor)
		}
		return n * 1000, nil
	}
	if strings.HasSuffix(waitFor, "min") {
		n, err := strconv.ParseInt(waitFor[:len(waitFor)-3], 0, 64)
		if err != nil {
			return 0, fmt.Errorf("--waitFor value %q is not a valid number", waitFor)
		}
		return n * 60 * 1000, nil
	}
	return 0, fmt.Errorf("--waitFor format is wrong. Accepted units are: milli, sec, min (e.g. 500milli, 30sec, 5min)")
}
