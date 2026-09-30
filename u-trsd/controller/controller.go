package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/u-trsd/usecase"
)

// Controller is an interface that defines the methods for handling requests.
type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	handler usecase.Handler
}

// NewController creates a new controller with the provided handler.
func NewController(handler usecase.Handler) Controller {
	return &controller{
		handler: handler,
	}
}

// Execute runs the handler associated with the controller using the provided context.
// It returns an error if the handler fails to run.
func (c *controller) Execute(ctx context.Context) error {
	err := c.handler.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run handler: %w", err)
	}

	return nil
}
