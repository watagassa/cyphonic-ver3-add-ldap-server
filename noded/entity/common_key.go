package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CommonKey struct {
	NodeID              []byte    `gorm:"column:node_id"`
	CommonKey           []byte    `gorm:"column:common_key"`
	CommonKeyCipherType int       `gorm:"column:common_key_cipher_type"`
	CommonKeyLength     int       `gorm:"column:common_key_length"`
	CommonKeyExpire     time.Time `gorm:"column:common_key_expire"`
}

// TypeCipherClass defines the class associated with crypt type in CYPHONIC packets.
type TypeCipherClass uint16

// TypeCipherClass known values.
const (
	AES256CBC TypeCipherClass = iota
	AES256CFB
	AES256OFB
	AES256CTR
	AES256GCM
)

// ExpireDate handles Common Key expiration.
type ExpireDate struct {
	Year  uint16
	Month uint8
	Day   uint8
}

// GenerateCommonKey generates a common key to encrypt the contents of the packet.
func GenerateCommonKey() ([]byte, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("missed new uuid random: %w", err)
	}

	commonKey := strings.ReplaceAll(u.String(), "-", "")

	return []byte(commonKey), nil
}
