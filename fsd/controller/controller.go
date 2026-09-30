package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/fsd/usecase"
)

type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	finalization usecase.Finalization
}

func NewController(finalization usecase.Finalization) Controller {
	return &controller{
		finalization: finalization,
	}
}

func (c *controller) Execute(ctx context.Context) error {
	return fmt.Errorf("can't call Finalization: %w", c.finalization.Finalization(ctx))
}
