package entity

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

type ProvisionResponse struct {
	BaseHeader         BaseHeader
	VirtualIPv4Address netip.Addr
	VirtualIPv6Address netip.Addr
	VirtualIPv4Prefix  uint16
	VirtualIPv6Prefix  uint16
	VIPVersion         VIPVersionClass
	L2Flag             L2FlagClass
	FQDNLength         uint16
	FQDN               []byte
}

type VIPVersionClass uint8

const (
	VIPVersionClassv4 VIPVersionClass = 4
	VIPVersionClassv6 VIPVersionClass = 6
)

type L2FlagClass uint8

const (
	L2FlagClassInactive L2FlagClass = 0
	L2FlagClassActive   L2FlagClass = 1
)

const (
	VirtualIPv4AddressSize      = 4
	VirtualIPv6AddressSize      = 16
	VirtualIPv4PrefixLengthSize = 2
	VirtualIPv6PrefixLengthSize = 2
	VirtualIPVersionSize        = 1
	L2FlagSize                  = 1
	FQDNLengthSize              = 2
)

const BaseProvisionResponseLength = BaseHeaderLength +
	VirtualIPv4AddressSize + VirtualIPv6AddressSize + VirtualIPv4PrefixLengthSize + VirtualIPv6PrefixLengthSize +
	VirtualIPVersionSize + L2FlagSize + FQDNLengthSize

func UnmarshalProvisionResponse(buffer []byte) (*ProvisionResponse, error) {
	if len(buffer) < BaseProvisionResponseLength {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBufferLength, len(buffer))
	}

	baseHeader := UnmarshalBaseHeader(buffer[:BaseHeaderLength])

	v4AddrStart := BaseHeaderLength
	v4AddrEnd := v4AddrStart + VirtualIPv4AddressSize
	v6AddrStart := v4AddrEnd
	v6AddrEnd := v6AddrStart + VirtualIPv6AddressSize
	v4PrefixStart := v6AddrEnd
	v4PrefixEnd := v4PrefixStart + VirtualIPv4PrefixLengthSize
	v6PrefixStart := v4PrefixEnd
	v6PrefixEnd := v6PrefixStart + VirtualIPv6PrefixLengthSize
	vipVersionStart := v6PrefixEnd
	vipVersionEnd := vipVersionStart + VirtualIPVersionSize
	l2FlagStart := vipVersionEnd
	l2FlagEnd := l2FlagStart + L2FlagSize
	fqdnLengthStart := l2FlagEnd
	fqdnLengthEnd := fqdnLengthStart + FQDNLengthSize
	fqdnStart := fqdnLengthEnd
	fqdnEnd := fqdnStart + int(binary.BigEndian.Uint16(buffer[fqdnLengthStart:fqdnLengthEnd]))

	return &ProvisionResponse{
		BaseHeader:         *baseHeader,
		VirtualIPv4Address: netip.AddrFrom4([4]byte(buffer[v4AddrStart:v4AddrEnd])),
		VirtualIPv6Address: netip.AddrFrom16([16]byte(buffer[v6AddrStart:v6AddrEnd])),
		VirtualIPv4Prefix:  binary.BigEndian.Uint16(buffer[v4PrefixStart:v4PrefixEnd]),
		VirtualIPv6Prefix:  binary.BigEndian.Uint16(buffer[v6PrefixStart:v6PrefixEnd]),
		VIPVersion:         VIPVersionClass(buffer[vipVersionStart]),
		L2Flag:             L2FlagClass(buffer[l2FlagStart]),
		FQDNLength:         binary.BigEndian.Uint16(buffer[fqdnLengthStart:fqdnLengthEnd]),
		FQDN:               buffer[fqdnStart:fqdnEnd],
	}, nil
}
