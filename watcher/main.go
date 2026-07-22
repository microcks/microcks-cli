package main

import (
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/watcher"
)

func main() {
	watchFile, err := config.DefaultLocalWatchPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wm, err := watcher.NewWatchManger(watchFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("[INFO] microcks-watcher started...")
	wm.Run()
}
