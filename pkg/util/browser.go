package util

import (
	"fmt"
	"os"

	"github.com/skratchdot/open-golang/open"
)

/*
LaunchBrowser opens the default system browser to the specified URL if conditions are met.
It skips opening the browser if the launchBrowser flag is false, if running in a CI environment,
or if the MICROCKS_NO_BROWSER environment variable is set.
*/
func LaunchBrowser(url string, launchBrowser bool) {
	if !launchBrowser {
		return
	}

	if os.Getenv("CI") != "" {
		//if in CI environment, do not attempt to open browser
		return
	}

	if os.Getenv("MICROCKS_NO_BROWSER") != "" {
		// if MICROCKS_NO_BROWSER environment variable is set, do not attempt to open browser
		return
	}

	fmt.Printf("Opening system default browser for %s\n", url)
	err := open.Start(url)
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}
