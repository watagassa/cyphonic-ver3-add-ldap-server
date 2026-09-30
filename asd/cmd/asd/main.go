// Package main contains the main work of the authentication service.
package main

import (
	"context"
	"os"

	"github.com/Pluslab/cyphonic/asd/injector"
)

func main() {
	injector := &injector.Injector{}
	ctx := context.Background()

	status := injector.Run(ctx)

	os.Exit(status)
}
