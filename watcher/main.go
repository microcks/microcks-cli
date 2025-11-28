package main

import (
	"fmt"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/watcher"
)

func main() {
	watchFile, err := config.DefaultLocalWatchPath()
	errors.CheckError(err)

	wm, err := watcher.NewWatchManger(watchFile)
	errors.CheckError(err)

	fmt.Println("[INFO] microcks-watcher started...")
	wm.Run()
}
