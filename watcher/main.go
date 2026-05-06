package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/microcks/microcks-cli/pkg/watcher"
)

func main() {
	watchFile, err := config.DefaultLocalWatchPath()
	errors.CheckError(err)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	wm, err := watcher.NewWatchManger(ctx, watchFile)
	errors.CheckError(err)

	fmt.Println("[INFO] microcks-watcher started...")
	wm.Run()
}
