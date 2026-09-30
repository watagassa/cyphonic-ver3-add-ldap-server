package entity

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

type ConnectionResponse struct {
	BaseHeader      BaseHeader
	NSIPv4Address   netip.Addr
	NSIPv6Address   netip.Addr
	NSPort          uint16
	Padding         uint16
	ExpireDate      ExpireDate
	CipherType      CipherClass
	CommonKeyLength uint16
	CommonKey       []byte
}

type CipherClass uint16

const (
	CipherClassAES256CBC CipherClass = 0
	CipherClassAES256CFB CipherClass = 1
	CipherClassAES256OFB CipherClass = 2
	CipherClassAES256CTR CipherClass = 3
	CipherClassAES256GCM CipherClass = 4
)

const (
	NSIPv4AddressSize   = 4
	NSIPv6AddressSize   = 16
	NSPortSize          = 2
	CipherTypeSize      = 2
	CommonKeyLengthSize = 2
)

const BaseConnectionResponseLength = BaseHeaderLength +
	NSIPv4AddressSize + NSIPv6AddressSize + NSPortSize + PaddingSize +
	ExpireDateSize + CipherTypeSize + CommonKeyLengthSize

func UnmarshalConnectionResponse(buffer []byte) (*ConnectionResponse, error) {
	if len(buffer) < BaseConnectionResponseLength {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBufferLength, len(buffer))
	}

	baseHeader := UnmarshalBaseHeader(buffer[:BaseHeaderLength])

	nsV4AddrStart := BaseHeaderLength
	nsV4AddrEnd := nsV4AddrStart + NSIPv4AddressSize
	nsV6AddrStart := nsV4AddrEnd
	nsV6AddrEnd := nsV6AddrStart + NSIPv6AddressSize
	nsPortStart := nsV6AddrEnd
	nsPortEnd := nsPortStart + NSPortSize
	expireDateStart := nsPortEnd + PaddingSize
	expireDateEnd := expireDateStart + ExpireDateSize
	cipherTypeStart := expireDateEnd
	cipherTypeEnd := cipherTypeStart + CipherTypeSize
	commonKeyLengthStart := cipherTypeEnd
	commonKeyLengthEnd := commonKeyLengthStart + CommonKeyLengthSize
	commonKeyStart := commonKeyLengthEnd
	commonKeyEnd := commonKeyStart + int(binary.BigEndian.Uint16(buffer[commonKeyLengthStart:commonKeyLengthEnd]))

	return &ConnectionResponse{
		BaseHeader:    *baseHeader,
		NSIPv4Address: netip.AddrFrom4([4]byte(buffer[nsV4AddrStart:nsV4AddrEnd])),
		NSIPv6Address: netip.AddrFrom16([16]byte(buffer[nsV6AddrStart:nsV6AddrEnd])),
		NSPort:        binary.BigEndian.Uint16(buffer[nsPortStart:nsPortEnd]),
		Padding:       binary.BigEndian.Uint16(buffer[nsPortEnd:expireDateStart]),
		ExpireDate: ExpireDate{
			Year:  binary.BigEndian.Uint16(buffer[expireDateStart : expireDateStart+YearSize]),
			Month: buffer[expireDateStart+YearSize],
			Day:   buffer[expireDateStart+YearSize+MonthSize],
		},
		CipherType:      CipherClass(buffer[cipherTypeStart]),
		CommonKeyLength: binary.BigEndian.Uint16(buffer[commonKeyLengthStart:commonKeyLengthEnd]),
		CommonKey:       buffer[commonKeyStart:commonKeyEnd],
	}, nil
}
