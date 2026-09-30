// Package main contains the main work of the finalization service.
package main

import (
	"context"
	"os"

	"github.com/Pluslab/cyphonic/fsd/injector"
)

func main() {
	injector := &injector.Injector{}
	ctx := context.Background()

	status := injector.Run(ctx)

	os.Exit(status)
}
