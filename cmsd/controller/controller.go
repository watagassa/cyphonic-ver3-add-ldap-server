package controller

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/cmsd/usecase"
)

type Controller interface {
	Execute(ctx context.Context) error
}

type controller struct {
	connectionManagement usecase.ConnectionManagement
}

func NewController(connectionManagement usecase.ConnectionManagement) Controller {
	return &controller{
		connectionManagement: connectionManagement,
	}
}

func (c *controller) Execute(ctx context.Context) error {
	return fmt.Errorf("can't call Connection Management: %w", c.connectionManagement.ConnectionManagement(ctx))
}
