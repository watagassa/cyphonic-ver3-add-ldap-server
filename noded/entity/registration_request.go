package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
)

type RegistrationRequest struct {
	BaseHeader          BaseHeader
	NodeIPv4Address     netip.Addr
	NodeIPv6Address     netip.Addr
	DaemonPort          uint16
	NotificationType    NotificationClass
	TunnelingProtocol   TunnelingProtocolClass
	InterfaceNameLength uint16
	InterfaceName       []byte
	HMAC                HMAC
}

type NotificationClass uint16

const (
	NotificationClassDefault NotificationClass = iota
	NotificationClassFirebaseCloudMessaging
	NotificationClassApplePushNotificationService
	NotificationClassWindowsNotificationHubs
)

type TunnelingProtocolClass uint16

const (
	TunnelingProtocolClassNone TunnelingProtocolClass = iota
	TunnelingProtocolClassUDP
	TunnelingProtocolClassQUIC
	TunnelingProtocolClassBoth
)

const (
	NodeIPv4AddressSize     = 4
	NodeIPv6AddressSize     = 16
	DaemonPortSize          = 2
	NotificationTypeSize    = 2
	TunnelingProtocolSize   = 2
	InterfaceNameLengthSize = 2
)

const daemonPort = 30000

const BaseRegistrationRequestLength = BaseHeaderLength +
	NodeIPv4AddressSize + NodeIPv6AddressSize + DaemonPortSize +
	NotificationTypeSize + TunnelingProtocolSize + InterfaceNameLengthSize + HMACLength

func GenerateRegistrationRequest(baseHeader BaseHeader, nodeIPv4Address, nodeIPv6Address netip.Addr, interfaceName string) *RegistrationRequest {
	BaseRegistrationRequestLength := BaseRegistrationRequestLength
	baseHeader.SerializeBaseHeader(TypeClassRegistrationRequest, StatusClassSuccess, uint16(BaseRegistrationRequestLength))
	// TODO: Daemonのポート番号を取得する必要がある
	// TODO: Interfaceに関しても同様
	return &RegistrationRequest{
		BaseHeader:          baseHeader,
		NodeIPv4Address:     nodeIPv4Address,
		NodeIPv6Address:     nodeIPv6Address,
		DaemonPort:          daemonPort,
		NotificationType:    NotificationClassDefault,
		TunnelingProtocol:   TunnelingProtocolClassUDP,
		InterfaceNameLength: uint16(len(interfaceName)),
		InterfaceName:       []byte(interfaceName),
		HMAC: HMAC{
			HMAC: make([]byte, HMACLength),
		},
	}
}

func (rreq RegistrationRequest) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	BaseHeaderBuf := rreq.BaseHeader.All()
	if _, err := buffer.Write(BaseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	if !rreq.NodeIPv4Address.IsValid() {
		rreq.NodeIPv4Address = ZeroAddr4()
	}

	if !rreq.NodeIPv6Address.IsValid() {
		rreq.NodeIPv6Address = ZeroAddr6()
	}

	fields := []interface{}{
		rreq.NodeIPv4Address.AsSlice(),
		rreq.NodeIPv6Address.AsSlice(),
		rreq.DaemonPort,
		rreq.NotificationType,
		rreq.TunnelingProtocol,
		rreq.InterfaceNameLength,
		rreq.InterfaceName,
		rreq.HMAC.HMAC,
	}

	for _, field := range fields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to write field: %w", err)
		}
	}

	return buffer.Bytes(), nil
}
