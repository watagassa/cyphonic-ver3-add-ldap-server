package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/psd/usecase"
)

type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	provision usecase.Provision
}

func NewController(provision usecase.Provision) Controller {
	return &controller{
		provision: provision,
	}
}

func (c *controller) Execute(ctx context.Context) error {
	return fmt.Errorf("can't call Provision: %w", c.provision.Provision(ctx))
}
