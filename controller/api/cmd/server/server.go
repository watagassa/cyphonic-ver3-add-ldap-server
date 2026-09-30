// Package main contains the main work of the cloud controller api service.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Pluslab/cyphonic/controller/api/injector"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Fprintln(os.Stdout, help())
		os.Exit(0)
	}

	ctx := context.Background()
	injector := &injector.Injector{}
	injector.Run(ctx)
}
