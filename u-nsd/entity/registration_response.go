package entity

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

type RegistrationResponse struct {
	BaseHeader    BaseHeader
	KeepAliveFlag KeepAliveFlagClass
	Padding       [3]byte
	HMAC          HMAC
}

type KeepAliveFlagClass uint8

const (
	KeepAliveNotRequired KeepAliveFlagClass = iota
	KeepAliveRequired
)

// Payload length in RegistrationResponse.
const (
	// KeepAliveFlagSize lengths (bytes).
	KeepAliveFlagSize = 1
	//
	// RegistrationResponsePaddingSize lengths (bytes).
	RegistrationResponsePaddingSize = 3
)

// BaseRegistrationResponseLength is a length of (base) body part of LoginRequest.
// Also, since the length of InterfaceName is variable length, when
// marshaling, calculate and add.
const BaseRegistrationResponseLength = BaseHeaderLength +
	KeepAliveFlagSize +
	RegistrationResponsePaddingSize +
	HMACLength

func GenerateRegisrationResponse(registrationRequest *RegistrationRequest, commonKey []byte, initiator *NodeAddress) (*RegistrationResponse, error) {
	registrationResponse := &RegistrationResponse{
		BaseHeader: BaseHeader{
			TransactionID:  registrationRequest.BaseHeader.TransactionID,
			Version:        registrationRequest.BaseHeader.Version,
			Type:           TypeClassRegistrationResponse,
			Status:         0,
			Count:          registrationRequest.BaseHeader.Count,
			SequenceNumber: registrationRequest.BaseHeader.SequenceNumber,
			MessageLength:  BaseRegistrationResponseLength + HMACLength,
			Token:          0,
			NextOpt:        0,
			ID:             registrationRequest.BaseHeader.ID,
		},
		KeepAliveFlag: KeepAliveRequired,
		Padding:       [3]byte{0, 0, 0},
		HMAC: []byte{
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
	}

	realIPv4addr, ok := netip.AddrFromSlice(initiator.RealIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert RealIPv4")
	}

	realIPv6addr, ok := netip.AddrFromSlice(initiator.RealIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert RealIPv6")
	}

	if (realIPv4addr.Compare(registrationRequest.NodeIPv4Address) == 0 ||
		realIPv6addr.Compare(registrationRequest.NodeIPv6Address) == 0) &&
		initiator.NATPort == int(registrationRequest.DaemonPort) {
		registrationResponse.KeepAliveFlag = KeepAliveNotRequired
	}

	return registrationResponse, nil
}

func (response *RegistrationResponse) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseRegistrationResponseLength)

	binary.BigEndian.PutUint32(buf[0:4], response.BaseHeader.TransactionID)
	buf[4] = response.BaseHeader.Version
	buf[5] = byte(response.BaseHeader.Type)
	buf[6] = byte(response.BaseHeader.Status)
	buf[7] = response.BaseHeader.Count
	binary.BigEndian.PutUint32(buf[8:12], response.BaseHeader.SequenceNumber)
	binary.BigEndian.PutUint16(buf[12:14], response.BaseHeader.MessageLength)
	buf[14] = response.BaseHeader.Token
	buf[15] = response.BaseHeader.NextOpt
	binary.BigEndian.PutUint64(buf[16:24], binary.BigEndian.Uint64(response.BaseHeader.ID[0:8]))
	binary.BigEndian.PutUint64(buf[24:32], binary.BigEndian.Uint64(response.BaseHeader.ID[8:16]))
	buf[32] = uint8(response.KeepAliveFlag)
	binary.BigEndian.PutUint16(buf[33:35], binary.BigEndian.Uint16(response.Padding[0:2]))
	buf[35] = response.Padding[2]
	binary.BigEndian.PutUint64(buf[36:44], binary.BigEndian.Uint64(response.HMAC[0:8]))
	binary.BigEndian.PutUint64(buf[44:52], binary.BigEndian.Uint64(response.HMAC[8:16]))

	return buf, nil
}

func (response *RegistrationResponse) SerializeBaseHeader(prevBaseHeader BaseHeader, typ TypeClass, st StatusClass) {
	response.BaseHeader.serializeTransactionID(prevBaseHeader)
	response.BaseHeader.serializeVersion()
	response.BaseHeader.serializeType(typ)
	response.BaseHeader.serializeStatus(st)
	response.BaseHeader.serializeCount()
	response.BaseHeader.serializeSequenceNumber()
	response.BaseHeader.serializeMessageLength(uint16(BaseRegistrationResponseLength))
	response.BaseHeader.serializeToken()
	response.BaseHeader.serializeNextOpt()
	response.BaseHeader.serializeID(prevBaseHeader.ID)
}

func (response *RegistrationResponse) SerializeKeepAliveFlag(kaFlag KeepAliveFlagClass) {
	response.KeepAliveFlag = kaFlag
}

func (response *RegistrationResponse) SerializePadding() {
	response.Padding = [3]byte{0, 0, 0}
}

func (response *RegistrationResponse) SerializeHMAC(hmac HMAC) {
	response.HMAC = hmac
}
