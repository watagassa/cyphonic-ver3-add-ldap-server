// Package controller assembles InputPort, OutputPort, and repository and executes InputPort.
package controller

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
)

// Healthz provides a health check endpoint.
func (c *Controller) Healthz(ctx context.Context) func(c echo.Context) error {
	return func(echo echo.Context) error {
		inputPort, err := c.newInputPort(echo)
		if err != nil {
			return fmt.Errorf("failed to return healthz: %w", err)
		}

		return fmt.Errorf("%w", inputPort.Healthz(ctx))
	}
}
