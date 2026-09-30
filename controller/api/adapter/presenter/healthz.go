// Package presenter implements the entity of OutputPort with technical elements.
package presenter

import (
	"fmt"
	"net/http"

	"github.com/Pluslab/cyphonic/controller/api/domain/message"
)

func (p *Presenter) OutputHealthz(msg message.SimpleResponse) error {
	return fmt.Errorf("%w", p.ctx.JSON(http.StatusOK, msg))
}
