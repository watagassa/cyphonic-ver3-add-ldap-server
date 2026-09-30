//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/asd/$GOPACKAGE
package infrastructure

import (
	"fmt"

	"github.com/google/uuid"
)

type idGenerator struct{}

func NewIDGenerator() *idGenerator {
	return &idGenerator{}
}

func (ig *idGenerator) GenerateNodeID(fqdn string, applicationID []byte) ([]byte, error) {
	fqdnBytes := []byte(fqdn)
	fqdnBytes = append(fqdnBytes, applicationID...)
	nodeID := uuid.NewSHA1(uuid.NameSpaceDNS, fqdnBytes)

	b, err := nodeID.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("cant't create node id: %w", err)
	}

	return b, nil
}
