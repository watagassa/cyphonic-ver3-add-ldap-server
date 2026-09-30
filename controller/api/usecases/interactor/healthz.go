// Package interactor defines an InputPort entity that receives request data and a processing flow that leaves out the technical elements.
package interactor

import (
	"context"
	"fmt"
)

func (i *Interactor) Healthz(ctx context.Context) error {
	return fmt.Errorf("%w", i.OutputPort.OutputSimpleMessage("health check endpoint"))
}
