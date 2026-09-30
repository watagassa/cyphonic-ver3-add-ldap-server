package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/asd/usecase"
)

type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	authentication usecase.Authentication
}

func NewController(authentication usecase.Authentication) Controller {
	return &controller{
		authentication: authentication,
	}
}

func (c *controller) Execute(ctx context.Context) error {
	return fmt.Errorf("can't call Authenticate: %w", c.authentication.Authenticate(ctx))
}
