package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/dsd/usecase"
)

var _ Controller = (*controller)(nil)

type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	handler usecase.Handler
}

func NewController(handler usecase.Handler) *controller {
	return &controller{
		handler: handler,
	}
}

func (c *controller) Execute(ctx context.Context) error {
	err := c.handler.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run handler: %w", err)
	}

	return nil
}
