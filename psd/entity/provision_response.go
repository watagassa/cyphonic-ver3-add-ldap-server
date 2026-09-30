package entity

import (
	"encoding/binary"
	"errors"
	"net"
	"net/netip"
)

// ProvisionResponse is Provision Response of a CYPHONIC packet.
type ProvisionResponse struct {
	BaseHeader         BaseHeader
	VirtualIPv4Address netip.Addr
	VirtualIPv6Address netip.Addr
	VirtualIPv4Prefix  uint16
	VirtualIPv6Prefix  uint16
	VIPVersion         VIPVersionClass
	L2Flag             L2FlagClass
	FQDNLength         uint16
	FQDN               string
}

type VIPVersionClass uint8

// VIPVersion known values.
const (
	VIPVersionClassv4 VIPVersionClass = 4
	VIPVersionClassv6 VIPVersionClass = 6
)

type L2FlagClass uint8

// L2Flag known values.
const (
	L2FlagClassInactive L2FlagClass = 0
	L2FlagClassActive   L2FlagClass = 1
)

// Payload length in ProvisionResponse.
const (
	// VirtualIPv4AddressSize is VIPv4 field lengths (bytes).
	VirtualIPv4AddressSize = 4

	// VirtualIPv6AddressSize is VIPv6 field length (bytes).
	VirtualIPv6AddressSize = 16

	// VirtualIPv4PrefixLengthSize is VIPv4 netmask field length
	VirtualIPv4PrefixLengthSize = 2

	// VirtualIPv6PrefixLengthSize is VIPv6 prefix field length
	VirtualIPv6PrefixLengthSize = 2

	// VirtualIPVersionSize is VIPversion field length
	VirtualIPVersionSize = 1

	// L2FlagSize is L2 Flag field length
	L2FlagSize = 1

	// FQDNLengthSize lengths (bytes).
	FQDNLengthSize = 2
)

// BaseProvisionResponseLength is length of (base) body part of ProvisionResponse.
// Also, since the length of FQDN is variable length, when
// marshaling, calculate and add.
const BaseProvisionResponseLength = BaseHeaderLen +
	VirtualIPv4AddressSize + VirtualIPv6AddressSize + VirtualIPv4PrefixLengthSize + VirtualIPv6PrefixLengthSize +
	VirtualIPVersionSize + L2FlagSize + FQDNLengthSize

func (response *ProvisionResponse) SerializeVirtualIP(vip4, vip6 netip.Prefix, vipVersion VIPVersionClass) {
	response.VirtualIPv4Address = vip4.Addr()
	response.VirtualIPv4Prefix = uint16(vip4.Bits())
	response.VirtualIPv6Address = vip6.Addr()
	response.VirtualIPv6Prefix = uint16(vip6.Bits())
	response.VIPVersion = vipVersion
}

func (response *ProvisionResponse) SerializeL2Flag() {
	response.L2Flag = L2FlagClassInactive // TODO : set L2Flag
}

func (response *ProvisionResponse) SerializeFQDN(fqdn string) {
	response.FQDN = fqdn
	response.FQDNLength = uint16(len(fqdn))
}

func (response *ProvisionResponse) SerializeBaseHeader(prevBaseHeader BaseHeader, typ TypeClass, st StatusClass) {
	response.BaseHeader.serializeTransactionID(prevBaseHeader)
	response.BaseHeader.serializeVersion()
	response.BaseHeader.serializeType(typ)
	response.BaseHeader.serializeStatus(st)
	response.BaseHeader.serializeCount()
	response.BaseHeader.serializeSequenceNumber(prevBaseHeader)
	response.BaseHeader.serializeMessageLength(uint16(BaseProvisionResponseLength + response.FQDNLength))
	response.BaseHeader.serializeToken()
	response.BaseHeader.serializeNextOpt()
	response.BaseHeader.serializeID(prevBaseHeader.ID)
}

func (response *ProvisionResponse) ChangeStatus(st StatusClass) {
	response.BaseHeader.serializeStatus(st)
}

func (response *ProvisionResponse) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseProvisionResponseLength, BaseProvisionResponseLength+int(response.FQDNLength))

	// Base Header
	binary.BigEndian.PutUint32(buf[0:4], response.BaseHeader.TransactionID)
	buf[4] = response.BaseHeader.Version
	buf[5] = response.BaseHeader.Type
	buf[6] = response.BaseHeader.Status
	buf[7] = response.BaseHeader.Count
	binary.BigEndian.PutUint32(buf[8:12], response.BaseHeader.SequenceNumber)
	binary.BigEndian.PutUint16(buf[12:14], response.BaseHeader.MessageLength)
	buf[14] = response.BaseHeader.Token
	buf[15] = response.BaseHeader.NextOpt
	binary.BigEndian.PutUint64(buf[16:24], binary.BigEndian.Uint64(response.BaseHeader.ID[0:8]))
	binary.BigEndian.PutUint64(buf[24:32], binary.BigEndian.Uint64(response.BaseHeader.ID[8:16]))
	// Provision Response Body
	binary.BigEndian.PutUint32(buf[32:36], binary.BigEndian.Uint32(net.IP(response.VirtualIPv4Address.AsSlice())))
	binary.BigEndian.PutUint64(buf[36:44], binary.BigEndian.Uint64(net.IP(response.VirtualIPv6Address.AsSlice()[0:8])))
	binary.BigEndian.PutUint64(buf[44:52], binary.BigEndian.Uint64(net.IP(response.VirtualIPv6Address.AsSlice()[8:16])))
	binary.BigEndian.PutUint16(buf[52:54], response.VirtualIPv4Prefix)
	binary.BigEndian.PutUint16(buf[54:56], response.VirtualIPv6Prefix)
	buf[56] = uint8(response.VIPVersion)
	buf[57] = uint8(response.L2Flag)
	binary.BigEndian.PutUint16(buf[58:60], response.FQDNLength)

	bfqdn := []byte(response.FQDN)
	if len(bfqdn) != int(response.FQDNLength) {
		return nil, errors.New("FQDN does not match FQDNLength")
	}

	buf = append(buf, bfqdn...)

	if len(buf) != int(response.BaseHeader.MessageLength) {
		return nil, errors.New("message does not match MessageLength")
	}

	return buf, nil
}
