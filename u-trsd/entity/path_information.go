package entity

import "net"

// PathInformation is a struct of CYPHONIC node to node path information.
type PathInformation struct {
	GeneratingFlag      bool
	InitiatorIPv4       net.IP
	InitiatorIPv6       net.IP
	InitiatorPort       uint16
	ResponderIPv4       net.IP
	ResponderIPv6       net.IP
	ResponderPort       uint16
	TunnelKey           []byte
	TunnelKeyCipherType uint16
	TunnelKeyLength     uint16
}

// PathInfoClass defines the class associated with a CYPHONIC packet flag.
// classes can be thought of as an array of parallel namespace trees.
type PathInfoClass uint8

// PathInfoClass known values.
const (
	PathInfoClassGeneratingFlag PathInfoClass = iota
	PathInfoClassInitiatorNodeIPv4
	PathInfoClassInitiatorNodeIPv6
	PathInfoClassInitiatorPort
	PathInfoClassResponderNodeIpv4
	PathInfoClassResponderNodeIpv6
	PathInfoClassResponderNodePort
	PathInfoClassTunnelKey
	PathInfoClassTunnelKeyCipherType
	PathInfoClassTunnelKeyLength
)

// GetPathInformationField returns the fields of PathInformation from the PathInformation struct.
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

// GeneratePathInformation creates a PathInformation struct from the provided values.
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
