package entity

import (
	"encoding/binary"
	"errors"
	"net/netip"
)

// ConnectionResponse is Connection Response of a CYPHONIC packet.
type ConnectionResponse struct {
	BaseHeader      BaseHeader
	NSIPv4Address   netip.Addr
	NSIPv6Address   netip.Addr
	NSPort          uint16
	Padding         uint16
	ExpireDate      ExpireDate
	CipherType      TypeCipherClass // future work (VIP version(8), L2Flag(8))
	CommonKeyLength uint16
	CommonKey       []byte
}

type ExpireDate struct {
	Year  uint16
	Month uint8
	Day   uint8
}

const (
	// YearSize is length of year (bytes).
	YearSize = 2

	// MonthSize is length of month (bytes).
	MonthSize = 1

	// DaySize is length of day (bytes).
	DaySize = 1
)

type TypeCipherClass uint16

const (
	AES256CBC TypeCipherClass = 0
	AES256CFB TypeCipherClass = 1
	AES256OFB TypeCipherClass = 2
	AES256CTR TypeCipherClass = 3
	AES256GCM TypeCipherClass = 4
)

// Payload length in LoginResponse.
const (
	// NSIPv4AddressSize lengths (bytes).
	NSIPv4AddressSize = 4

	// NSIPv4AddressSize lengths (bytes).
	NSIPv6AddressSize = 16

	// NSPortSize lengths (bytes).
	NSPortSize = 2

	// ExpireDateSize lengths (bytes).
	ExpireDateSize = YearSize + MonthSize + DaySize

	// CommonKeySize lengths (bytes).
	CipherTypeSize = 2

	// CommonKeyLengthSize lengths (bytes).
	CommonKeyLengthSize = 2
)

// BaseConnectionResponseLength is length of (base) body part of ConnectionResponse.
// Also, since the length of CommonKey is variable length, when
// marshaling, calculate and add.
const BaseConnectionResponseLength = BaseHeaderLen +
	NSIPv4AddressSize + NSIPv6AddressSize +
	NSPortSize + PaddingSize +
	ExpireDateSize + CipherTypeSize + CommonKeyLengthSize

func (response *ConnectionResponse) SerializeNSInformation(nsInfo *NotificationServiceInfomation) {
	response.NSIPv4Address = nsInfo.IPv4
	response.NSIPv6Address = nsInfo.IPv6
	response.NSPort = nsInfo.Port
}

func (response *ConnectionResponse) SerializeCommonKey(commonKey *CommonKey) {
	response.ExpireDate.Year = commonKey.ExpireDate.Year
	response.ExpireDate.Month = commonKey.ExpireDate.Month
	response.ExpireDate.Day = commonKey.ExpireDate.Day
	response.CipherType = commonKey.CipherType

	response.CommonKey = []byte(commonKey.CommonKey)
	response.CommonKeyLength = uint16(len(commonKey.CommonKey))
}

func (response *ConnectionResponse) SerializeBaseHeader(prevBaseHeader BaseHeader, typ TypeClass, st StatusClass) {
	response.BaseHeader.serializeTransactionID(prevBaseHeader)
	response.BaseHeader.serializeVersion()
	response.BaseHeader.serializeType(typ)
	response.BaseHeader.serializeStatus(st)
	response.BaseHeader.serializeCount()
	response.BaseHeader.serializeSequenceNumber(prevBaseHeader)
	response.BaseHeader.serializeMessageLength(uint16(BaseConnectionResponseLength + response.CommonKeyLength))
	response.BaseHeader.serializeToken()
	response.BaseHeader.serializeNextOpt()
	response.BaseHeader.serializeID(prevBaseHeader.ID)
}

func (response *ConnectionResponse) ChangeStatus(st StatusClass) {
	response.BaseHeader.serializeStatus(st)
}

func (response *ConnectionResponse) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseConnectionResponseLength, BaseConnectionResponseLength+response.CommonKeyLength)

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

	// Connection Response body
	binary.BigEndian.PutUint32(buf[32:36], binary.BigEndian.Uint32(response.NSIPv4Address.AsSlice()))
	binary.BigEndian.PutUint64(buf[36:44], binary.BigEndian.Uint64(response.NSIPv6Address.AsSlice()[0:8]))
	binary.BigEndian.PutUint64(buf[44:52], binary.BigEndian.Uint64(response.NSIPv6Address.AsSlice()[8:16]))
	binary.BigEndian.PutUint16(buf[52:54], response.NSPort)
	binary.BigEndian.PutUint16(buf[54:56], response.Padding)
	binary.BigEndian.PutUint16(buf[56:58], response.ExpireDate.Year)
	buf[58] = response.ExpireDate.Month
	buf[59] = response.ExpireDate.Day
	binary.BigEndian.PutUint16(buf[60:62], uint16(response.CipherType))
	binary.BigEndian.PutUint16(buf[62:64], response.CommonKeyLength)

	if len(response.CommonKey) != int(response.CommonKeyLength) {
		return nil, errors.New("common key does not match common key length")
	}

	buf = append(buf, response.CommonKey...)

	return buf, nil
}
