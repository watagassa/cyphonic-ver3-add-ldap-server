// Package main contains the main work of the node management service.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Pluslab/cyphonic/u-nsd/injector"
)

func main() {
	injector := &injector.Injector{}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	status := injector.Run(ctx)

	os.Exit(status)
}
