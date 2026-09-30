package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, httpServer *http.Server) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("failed to listen and serve: %w", err)
	})
	errGroup.Go(func() error {
		<-ctx.Done()

		return httpServer.Shutdown(ctx)
	})

	config.LogInfo("daemon API server is running...")

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
