// Package main contains the main work of the connection management service.
package main

import (
	"context"
	"os"

	"github.com/Pluslab/cyphonic/cmsd/injector"
)

func main() {
	injector := &injector.Injector{}
	ctx := context.Background()

	status := injector.Run(ctx)

	os.Exit(status)
}
