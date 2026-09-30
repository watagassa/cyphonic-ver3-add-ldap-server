package entity

import (
	"net"
	"time"
)

type PathInformation struct {
	PathID                []byte    `json:"path_id"`
	GeneratingFlag        bool      `json:"generating_flag"`
	InitiatorIPv4         net.IP    `json:"initiator_ipv4"`
	InitiatorIPv6         net.IP    `json:"initiator_ipv6"`
	InitiatorPort         uint16    `json:"initiator_port"`
	InitiatorConnectionID []byte    `json:"initiator_connection_id"`
	ResponderIPv4         net.IP    `json:"responder_ipv4"`
	ResponderIPv6         net.IP    `json:"responder_ipv6"`
	ResponderPort         uint16    `json:"responder_port"`
	ResponderConnectionID []byte    `json:"responder_connection_id"`
	TunnelKey             []byte    `json:"tunnel_key"`
	TunnelKeyCipherType   uint16    `json:"tunnel_key_cipher_type"`
	TunnelKeyLength       uint16    `json:"tunnel_key_length"`
	TunnelKeyExpire       time.Time `json:"tunnel_key_expire"`
	CreatedAt             time.Time `gorm:"column:created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at"`
}

// PathInfoClass defines the class associated with a CYPHONIC packet flag.
// classes can be thought of as an array of parallel namespace trees.
type PathInfoClass uint8

// PathInfoClass known values.
const (
	PathInfoClassGeneratingFlag      PathInfoClass = 0
	PathInfoClassInitiatorNodeIPv4   PathInfoClass = 1
	PathInfoClassInitiatorNodeIPv6   PathInfoClass = 2
	PathInfoClassInitiatorPort       PathInfoClass = 3
	PathInfoClassResponderNodeIpv4   PathInfoClass = 4
	PathInfoClassResponderNodeIpv6   PathInfoClass = 5
	PathInfoClassResponderNodePort   PathInfoClass = 6
	PathInfoClassTunnelKey           PathInfoClass = 7
	PathInfoClassTunnelKeyCipherType PathInfoClass = 8
	PathInfoClassTunnelKeyLength     PathInfoClass = 9
)

func GetPathInformationField() []string {
	return []string{
		"generatingFlag",
		"initiatorIPv4",
		"initiatorIPv6",
		"initiatorPort",
		"responderIPv4",
		"responderIPv6",
		"responderPort",
		"tunnelKey",
		"tunnelKeyCipherType",
		"tunnelKeyLength",
	}
}

func GeneratePathInformation(value []interface{}) *PathInformation {
	generatingFlag, ok1 := value[0].(bool)
	if !ok1 {
		return nil
	}

	initiatorIPv4, ok2 := value[1].(net.IP)
	if !ok2 {
		return nil
	}

	initiatorIPv6, ok3 := value[2].(net.IP)
	if !ok3 {
		return nil
	}

	initiatorPort, ok4 := value[3].(uint16)
	if !ok4 {
		return nil
	}

	responderIPv4, ok5 := value[4].(net.IP)
	if !ok5 {
		return nil
	}

	responderIPv6, ok6 := value[5].(net.IP)
	if !ok6 {
		return nil
	}

	responderPort, ok7 := value[6].(uint16)
	if !ok7 {
		return nil
	}

	tunnelKey, ok8 := value[7].([]byte)
	if !ok8 {
		return nil
	}

	tunnelKeyCipherType, ok9 := value[8].(uint16)
	if !ok9 {
		return nil
	}

	tunnelKeyLength, ok10 := value[9].(uint16)
	if !ok10 {
		return nil
	}

	return &PathInformation{
		GeneratingFlag:      generatingFlag,
		InitiatorIPv4:       initiatorIPv4,
		InitiatorIPv6:       initiatorIPv6,
		InitiatorPort:       initiatorPort,
		ResponderIPv4:       responderIPv4,
		ResponderIPv6:       responderIPv6,
		ResponderPort:       responderPort,
		TunnelKey:           tunnelKey,
		TunnelKeyCipherType: tunnelKeyCipherType,
		TunnelKeyLength:     tunnelKeyLength,
	}
}
