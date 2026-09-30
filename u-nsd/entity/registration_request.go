package entity

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
)

// RegistrationRequest is ...
type RegistrationRequest struct {
	BaseHeader          BaseHeader
	NodeIPv4Address     netip.Addr
	NodeIPv6Address     netip.Addr
	DaemonPort          uint16
	NotificationType    TypeNotificationClass
	TunnelingProtocol   TunnelingProtocolClass
	InterfaceNameLength uint16
	InterfaceName       InterfaceName
	HMAC                HMAC
}

type TypeNotificationClass uint16

type TunnelingProtocolClass uint16

const (
	TypeAllowUDP        TunnelingProtocolClass = 1
	TypeAllowQUIC       TunnelingProtocolClass = 2
	TypeAllowUDPAndQUIC TunnelingProtocolClass = 3
)

type InterfaceName []byte

// Payload length in RegistrationRequest.
const (
	// NodeIPv4AddressSize lengths (bytes).
	NodeIPv4AddressSize = 4

	// NodeIPv6AddressSize lengths (bytes).
	NodeIPv6AddressSize = 16

	// DaemonPortSize lengths (bytes).
	DaemonPortSize = 2

	// NotificationTypeSize lengths (bytes).
	NotificationTypeSize = 2

	// TunnelingProtocolSize lengths (bytes).
	TunnelingProtocolSize = 2

	// InterfaceNameLengthSize lengths (bytes).
	InterfaceNameLengthSize = 2
)

const (
	DaemonPortOffset          = NodeIPv4AddressSize + NodeIPv6AddressSize
	NotificationTypeOffset    = DaemonPortOffset + DaemonPortSize
	TunnelingProtocolOffset   = NotificationTypeOffset + NotificationTypeSize
	InterfaceNameLengthOffset = TunnelingProtocolOffset + TunnelingProtocolSize
	InterfaceNameOffset       = InterfaceNameLengthOffset + InterfaceNameLengthSize
)

// BaseRegistrationRequestLength is a length of (base) body part of LoginRequest.
// Also, since the length of InterfaceName and HMACLength is variable length, when
// marshaling, calculate and add.
const BaseRegistrationRequestLength = BaseHeaderLength +
	NodeIPv4AddressSize +
	NodeIPv6AddressSize +
	DaemonPortSize +
	NotificationTypeSize +
	TunnelingProtocolSize +
	InterfaceNameLengthSize

// ParseRegistrationRequest parses the raw data of a UDP packet and determines its type.
func ParseRegistrationRequest(packet *BasePacket, commonKey []byte) (*RegistrationRequest, error) {
	var nodeIPv4Address netip.Addr
	if err := nodeIPv4Address.UnmarshalBinary(packet.Payload[0:NodeIPv4AddressSize]); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node IPv4 address: %w", err)
	}

	var nodeIPv6Address netip.Addr
	if err := nodeIPv6Address.UnmarshalBinary(packet.Payload[NodeIPv4AddressSize : NodeIPv4AddressSize+NodeIPv6AddressSize]); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node IPv6 address: %w", err)
	}

	registrationRequest := &RegistrationRequest{
		BaseHeader:          packet.BaseHeader,
		NodeIPv4Address:     nodeIPv4Address,
		NodeIPv6Address:     nodeIPv6Address,
		DaemonPort:          binary.BigEndian.Uint16(packet.Payload[DaemonPortOffset:NotificationTypeOffset]),
		NotificationType:    TypeNotificationClass(binary.BigEndian.Uint16(packet.Payload[NotificationTypeOffset:TunnelingProtocolOffset])),
		TunnelingProtocol:   TunnelingProtocolClass(binary.BigEndian.Uint16(packet.Payload[TunnelingProtocolOffset:InterfaceNameLengthOffset])),
		InterfaceNameLength: binary.BigEndian.Uint16(packet.Payload[InterfaceNameLengthOffset:InterfaceNameOffset]),
		InterfaceName:       []byte{},
		HMAC:                packet.HMAC,
	}

	hmacOffset := InterfaceNameOffset + registrationRequest.InterfaceNameLength
	registrationRequest.InterfaceName = packet.Payload[InterfaceNameOffset:hmacOffset]

	// Check if the payload length matches the expected length
	expectedLength := BaseRegistrationRequestLength +
		int(registrationRequest.InterfaceNameLength) -
		BaseHeaderLength

	if len(packet.Payload) != expectedLength {
		return nil, fmt.Errorf("invalid packet length: expected %d, got %d", expectedLength, len(packet.Payload))
	}

	return registrationRequest, nil
}

func (response *RegistrationRequest) ConvertBytes() ([]byte, error) {
	buf := make([]byte, BaseRegistrationRequestLength,
		BaseRegistrationRequestLength+HMACLength+response.InterfaceNameLength)

	nodeIPv4AddressBin, err := response.NodeIPv4Address.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NodeIPv4Address: %w", err)
	}

	nodeIPv6AddressBin, err := response.NodeIPv6Address.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal NodeIPv6Address: %w", err)
	}

	// Base Header
	binary.BigEndian.PutUint32(buf[0:4], response.BaseHeader.TransactionID)
	buf[4] = response.BaseHeader.Version
	buf[5] = uint8(response.BaseHeader.Type)
	buf[6] = byte(response.BaseHeader.Status)
	buf[7] = response.BaseHeader.Count
	binary.BigEndian.PutUint32(buf[8:12], response.BaseHeader.SequenceNumber)
	binary.BigEndian.PutUint16(buf[12:14], response.BaseHeader.MessageLength)
	buf[14] = response.BaseHeader.Token
	buf[15] = response.BaseHeader.NextOpt
	binary.BigEndian.PutUint64(buf[16:24], binary.BigEndian.Uint64(response.BaseHeader.ID[0:8]))
	binary.BigEndian.PutUint64(buf[24:32], binary.BigEndian.Uint64(response.BaseHeader.ID[8:16]))
	// Registration Request Body
	binary.BigEndian.PutUint32(buf[32:36], binary.BigEndian.Uint32(nodeIPv4AddressBin))
	binary.BigEndian.PutUint64(buf[36:44], binary.BigEndian.Uint64(nodeIPv6AddressBin[0:8]))
	binary.BigEndian.PutUint64(buf[44:52], binary.BigEndian.Uint64(nodeIPv6AddressBin[8:16]))

	binary.BigEndian.PutUint16(buf[52:54], response.DaemonPort)
	binary.BigEndian.PutUint16(buf[54:56], uint16(response.NotificationType))
	binary.BigEndian.PutUint16(buf[56:58], uint16(response.TunnelingProtocol))
	binary.BigEndian.PutUint16(buf[58:60], response.InterfaceNameLength)

	if len(response.InterfaceName) != int(response.InterfaceNameLength) {
		return nil, errors.New("InterfaceName does not match InterfaceName length")
	}

	buf = append(buf, response.InterfaceName...)

	binary.BigEndian.PutUint64(
		buf[60+int(response.InterfaceNameLength):68+int(response.InterfaceNameLength)],
		binary.BigEndian.Uint64(response.HMAC[0:8]))
	binary.BigEndian.PutUint64(
		buf[68+int(response.InterfaceNameLength):76+int(response.InterfaceNameLength)],
		binary.BigEndian.Uint64(response.HMAC[8:16]))

	return buf, nil
}
