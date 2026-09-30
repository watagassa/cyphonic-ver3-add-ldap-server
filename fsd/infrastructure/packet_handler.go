//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/fsd/$GOPACKAGE
package infrastructure

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/Pluslab/cyphonic/fsd/entity"
	"github.com/Pluslab/cyphonic/fsd/usecase/repository"
)

type packetHandler struct{}

func NewPacketHandler() repository.PacketHandler {
	return &packetHandler{}
}

func (ph *packetHandler) Unmarshal(ctx context.Context, buf []byte, length int) (*entity.FinalizationRequest, bool, error) {
	if length < entity.BaseFinalizationRequestLength {
		return nil, false, fmt.Errorf("invalid finalization request packet length")
	}

	shouldReadPacket := false

	if buf[5] == uint8(entity.TypeClassFinalizationRequest) {
		req := entity.FinalizationRequest{
			BaseHeader: entity.BaseHeader{
				TransactionID:  binary.BigEndian.Uint32(buf[0:4]),
				Version:        buf[4],
				Type:           entity.TypeClass(buf[5]),
				Status:         entity.StatusClass(buf[6]),
				Count:          buf[7],
				SequenceNumber: binary.BigEndian.Uint32(buf[8:12]),
				MessageLength:  binary.BigEndian.Uint16(buf[12:14]),
				Token:          buf[14],
				NextOpt:        buf[15],
				ID:             buf[16:32],
			},
		}

		return &req, shouldReadPacket, nil
	}

	return nil, false, fmt.Errorf("invalid packet type of finalization request")
}

func (ph *packetHandler) GenerateFinalizationResponse(baseHeader entity.BaseHeader) (*entity.FinalizationResponse, error) {
	baseHeader.SerializeBaseHeader(entity.TypeClassAck, entity.StatusClassSuccess, uint16(entity.BaseFinalizationResponseLength))

	return &entity.FinalizationResponse{
		BaseHeader: baseHeader,
	}, nil
}
