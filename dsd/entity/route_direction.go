package entity

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
	"time"
)

type RouteDirection struct {
	BaseHeader               BaseHeader
	PathID                   ID
	ProcessCode              ProcessCodeClass
	TunnelConnType           TunnelConnectionType
	GeneralNodeFlag          GeneralNodeFlagClass
	TRSPort                  uint16
	TRSIPv4                  [4]byte
	TRSIPv6                  [16]byte
	InitiatorVirtualIPv4     [4]byte
	InitiatorVirtualIPv6     [16]byte
	ResponderVirtualIPv4     [4]byte
	ResponderVirtualIPv6     [16]byte
	NATInitiatorPort         uint16
	NATResponderPort         uint16
	InitiatorInterfaceNumber uint8
	ResponderInterfaceNumber uint8
	TunnelKeyCipherType      TypeCipherClass
	TunnelKeyLength          uint16
	TemporaryKeyLength       uint16
	ExpireDate               ExpireDate
	FQDNInitiatorLength      uint16
	FQDNResponderLength      uint16
	NATInitiatorIPv4         [4]byte
	NATInitiatorIPv6         [16]byte
	NATResponderIPv4         [4]byte
	NATResponderIPv6         [16]byte
	InitiatorRealIPv4        [4]byte
	InitiatorRealIPv6        [16]byte
	ResponderRealIPv4        [4]byte
	ResponderRealIPv6        [16]byte
	TunnelKey                []byte
	TemporaryKey             []byte
	InitiatorFQDN            []byte
	ResponderFQDN            []byte
}

type ProcessCodeClass uint16

const (
	TunnelRequestToResponder    ProcessCodeClass = 1
	TunnelRequestToNATResponder ProcessCodeClass = 2
	TunnelRequestToTRS          ProcessCodeClass = 3
)

type GeneralNodeFlagClass uint16

const (
	CYPHONICNode GeneralNodeFlagClass = 0
	GeneralNode  GeneralNodeFlagClass = 1
)

const (
	BaseRouteDirectionLength = 216
	IDLength                 = 16
)

func GenerateRouteDirection(
	baseHeader *BaseHeader,
	pathID *ID,
	processCode ProcessCodeClass,
	tunnelKey []byte,
	initiatorNodeInformation, responderNodeInformation *NodeInformation,
	initiatorNodeAddress, responderNodeAddress *NodeAddress,
	trsInformation *TunnelRelayService,
	trsFQDN string,
) (*RouteDirection, error) {
	temporaryKey, err := GenerateCommonKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate temporary key: %w", err)
	}

	expireDate := time.Now().AddDate(0, 1, 0)

	trsIPv4, ok := netip.AddrFromSlice(trsInformation.TunnleRelayServiceIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert TRS IPv4 address: invalid address length")
	}

	trsIPv6, ok := netip.AddrFromSlice(trsInformation.TunnleRelayServiceIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert TRS IPv6 address")
	}

	initiatorVirtualIPv4, ok := netip.AddrFromSlice(initiatorNodeInformation.VirtualIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator virtual IPv4 address")
	}

	initiatorVirtualIPv6, ok := netip.AddrFromSlice(initiatorNodeInformation.VirtualIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator virtual IPv6 address")
	}

	responderVirtualIPv4, ok := netip.AddrFromSlice(responderNodeInformation.VirtualIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder virtual IPv4 address")
	}

	responderVirtualIPv6, ok := netip.AddrFromSlice(responderNodeInformation.VirtualIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder virtual IPv6 address")
	}

	initiatorNodeNATIPv4, ok := netip.AddrFromSlice(initiatorNodeAddress.NATIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator NAT IPv4 address")
	}

	initiatorNodeNATIPv6, ok := netip.AddrFromSlice(initiatorNodeAddress.NATIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator NAT IPv6 address")
	}

	responderNodeNATIPv4, ok := netip.AddrFromSlice(responderNodeAddress.NATIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder NAT IPv4 address")
	}

	responderNodeNATIPv6, ok := netip.AddrFromSlice(responderNodeAddress.NATIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder NAT IPv6 address")
	}

	initiatorNodeRealIPv4, ok := netip.AddrFromSlice(initiatorNodeAddress.RealIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator real IPv4 address")
	}

	initiatorNodeRealIPv6, ok := netip.AddrFromSlice(initiatorNodeAddress.RealIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert initiator real IPv6 address")
	}

	responderNodeRealIPv4, ok := netip.AddrFromSlice(responderNodeAddress.RealIPv4)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder real IPv4 address")
	}

	responderNodeRealIPv6, ok := netip.AddrFromSlice(responderNodeAddress.RealIPv6)
	if !ok {
		return nil, fmt.Errorf("failed to convert responder real IPv6 address")
	}

	baseHeader.MessageLength = BaseRouteDirectionLength + uint16(len(tunnelKey)) + uint16(len(temporaryKey)) + uint16(len(initiatorNodeInformation.FQDN)) + uint16(len(responderNodeInformation.FQDN))
	baseHeader.Type = TypeClassRouteDirectionToResponder
	// Set Responder Node ID to BaseHeader.ID
	baseHeader.ID = responderNodeInformation.NodeID

	return &RouteDirection{
		BaseHeader:               *baseHeader,
		PathID:                   *pathID,
		ProcessCode:              processCode,
		TunnelConnType:           TunnelConnectionUDP, // FIXME: set tunnel connection type
		GeneralNodeFlag:          CYPHONICNode,        // FIXME: set node flag
		TRSPort:                  uint16(trsInformation.TunnleRelayServicePort),
		TRSIPv4:                  trsIPv4.As4(),
		TRSIPv6:                  trsIPv6.As16(),
		InitiatorVirtualIPv4:     initiatorVirtualIPv4.As4(),
		InitiatorVirtualIPv6:     initiatorVirtualIPv6.As16(),
		ResponderVirtualIPv4:     responderVirtualIPv4.As4(),
		ResponderVirtualIPv6:     responderVirtualIPv6.As16(),
		NATInitiatorPort:         uint16(initiatorNodeAddress.NATPort),
		NATResponderPort:         uint16(responderNodeAddress.NATPort),
		InitiatorInterfaceNumber: 1,         // FIXME: set interface number
		ResponderInterfaceNumber: 1,         // FIXME: set interface number
		TunnelKeyCipherType:      AES256CBC, // FIXME: set cipher type
		TunnelKeyLength:          uint16(len(tunnelKey)),
		TemporaryKeyLength:       uint16(len(temporaryKey)),
		ExpireDate: ExpireDate{
			Year:  uint16(expireDate.Year()),
			Month: uint8(expireDate.Month()),
			Day:   uint8(expireDate.Day()),
		},
		FQDNInitiatorLength: uint16(len(initiatorNodeInformation.FQDN)),
		FQDNResponderLength: uint16(len(responderNodeInformation.FQDN)),
		NATInitiatorIPv4:    initiatorNodeNATIPv4.As4(),
		NATInitiatorIPv6:    initiatorNodeNATIPv6.As16(),
		NATResponderIPv4:    responderNodeNATIPv4.As4(),
		NATResponderIPv6:    responderNodeNATIPv6.As16(),
		InitiatorRealIPv4:   initiatorNodeRealIPv4.As4(),
		InitiatorRealIPv6:   initiatorNodeRealIPv6.As16(),
		ResponderRealIPv4:   responderNodeRealIPv4.As4(),
		ResponderRealIPv6:   responderNodeRealIPv6.As16(),
		TunnelKey:           tunnelKey,
		TemporaryKey:        temporaryKey,
		InitiatorFQDN:       []byte(initiatorNodeInformation.FQDN),
		ResponderFQDN:       []byte(responderNodeInformation.FQDN),
	}, nil
}

func (r *RouteDirection) Marshal() ([]byte, error) {
	data := make([]byte, 0)
	buffer := bytes.NewBuffer(data)

	baseHeaderBuf := r.BaseHeader.Serialize()
	if _, err := buffer.Write(baseHeaderBuf); err != nil {
		return nil, fmt.Errorf("failed to write base header: %w", err)
	}

	routeDirectionFields := []interface{}{
		r.PathID, r.ProcessCode, r.TunnelConnType, r.GeneralNodeFlag, r.TRSPort,
		r.TRSIPv4, r.TRSIPv6, r.InitiatorVirtualIPv4, r.InitiatorVirtualIPv6,
		r.ResponderVirtualIPv4, r.ResponderVirtualIPv6, r.NATInitiatorPort,
		r.NATResponderPort, r.InitiatorInterfaceNumber, r.ResponderInterfaceNumber,
		r.TunnelKeyCipherType, r.TunnelKeyLength, r.TemporaryKeyLength, r.ExpireDate,
		r.FQDNInitiatorLength, r.FQDNResponderLength, r.NATInitiatorIPv4,
		r.NATInitiatorIPv6, r.NATResponderIPv4, r.NATResponderIPv6, r.InitiatorRealIPv4,
		r.InitiatorRealIPv6, r.ResponderRealIPv4, r.ResponderRealIPv6,
	}

	for _, field := range routeDirectionFields {
		if err := binary.Write(buffer, binary.BigEndian, field); err != nil {
			return nil, fmt.Errorf("failed to marshal route direction: %w", err)
		}
	}

	buffer.Write(r.TunnelKey)
	buffer.Write(r.TemporaryKey)
	buffer.Write(r.InitiatorFQDN)
	buffer.Write(r.ResponderFQDN)

	return buffer.Bytes(), nil
}

func ParseRouteDirection(packet *BasePacket) (*RouteDirection, error) {
	// Check if the payload length is more than the minimum packetsize
	if len(packet.Payload) < (BaseRouteDirectionLength - BaseHeaderLength) {
		return nil, fmt.Errorf("invalid packet length: %d", len(packet.Payload))
	}

	routeDirection := &RouteDirection{
		BaseHeader:      packet.BaseHeader,
		PathID:          ID(packet.Payload[0:IDLength]),
		ProcessCode:     ProcessCodeClass(binary.BigEndian.Uint16(packet.Payload[IDLength : IDLength+2])),
		TunnelConnType:  TunnelConnectionType(binary.BigEndian.Uint16(packet.Payload[IDLength+2 : IDLength+4])),
		GeneralNodeFlag: GeneralNodeFlagClass(binary.BigEndian.Uint16(packet.Payload[IDLength+4 : IDLength+6])),
		TRSPort:         binary.BigEndian.Uint16(packet.Payload[IDLength+6 : IDLength+8]),
		TRSIPv4:         [4]byte{packet.Payload[IDLength+8], packet.Payload[IDLength+9], packet.Payload[IDLength+10], packet.Payload[IDLength+11]},
		TRSIPv6: [16]byte{
			packet.Payload[IDLength+12], packet.Payload[IDLength+13], packet.Payload[IDLength+14], packet.Payload[IDLength+15],
			packet.Payload[IDLength+16], packet.Payload[IDLength+17], packet.Payload[IDLength+18], packet.Payload[IDLength+19],
			packet.Payload[IDLength+20], packet.Payload[IDLength+21], packet.Payload[IDLength+22], packet.Payload[IDLength+23],
			packet.Payload[IDLength+24], packet.Payload[IDLength+25], packet.Payload[IDLength+26], packet.Payload[IDLength+27],
		},
		InitiatorVirtualIPv4: [4]byte{packet.Payload[IDLength+28], packet.Payload[IDLength+29], packet.Payload[IDLength+30], packet.Payload[IDLength+31]},
		InitiatorVirtualIPv6: [16]byte{
			packet.Payload[IDLength+32], packet.Payload[IDLength+33], packet.Payload[IDLength+34], packet.Payload[IDLength+35],
			packet.Payload[IDLength+36], packet.Payload[IDLength+37], packet.Payload[IDLength+38], packet.Payload[IDLength+39],
			packet.Payload[IDLength+40], packet.Payload[IDLength+41], packet.Payload[IDLength+42], packet.Payload[IDLength+43],
			packet.Payload[IDLength+44], packet.Payload[IDLength+45], packet.Payload[IDLength+46], packet.Payload[IDLength+47],
		},
		ResponderVirtualIPv4: [4]byte{packet.Payload[IDLength+48], packet.Payload[IDLength+49], packet.Payload[IDLength+50], packet.Payload[IDLength+51]},
		ResponderVirtualIPv6: [16]byte{
			packet.Payload[IDLength+52], packet.Payload[IDLength+53], packet.Payload[IDLength+54], packet.Payload[IDLength+55],
			packet.Payload[IDLength+56], packet.Payload[IDLength+57], packet.Payload[IDLength+58], packet.Payload[IDLength+59],
			packet.Payload[IDLength+60], packet.Payload[IDLength+61], packet.Payload[IDLength+62], packet.Payload[IDLength+63],
			packet.Payload[IDLength+64], packet.Payload[IDLength+65], packet.Payload[IDLength+66], packet.Payload[IDLength+67],
		},
		NATInitiatorPort:         binary.BigEndian.Uint16(packet.Payload[IDLength+68 : IDLength+70]),
		NATResponderPort:         binary.BigEndian.Uint16(packet.Payload[IDLength+70 : IDLength+72]),
		InitiatorInterfaceNumber: packet.Payload[IDLength+72],
		ResponderInterfaceNumber: packet.Payload[IDLength+73],
		TunnelKeyCipherType:      TypeCipherClass(binary.BigEndian.Uint16(packet.Payload[IDLength+74 : IDLength+76])),
		TunnelKeyLength:          binary.BigEndian.Uint16(packet.Payload[IDLength+76 : IDLength+78]),
		TemporaryKeyLength:       binary.BigEndian.Uint16(packet.Payload[IDLength+78 : IDLength+80]),
		ExpireDate: ExpireDate{
			Year:  binary.BigEndian.Uint16(packet.Payload[IDLength+80 : IDLength+82]),
			Month: packet.Payload[IDLength+82],
			Day:   packet.Payload[IDLength+83],
		},
		FQDNInitiatorLength: uint16(binary.BigEndian.Uint16(packet.Payload[IDLength+84 : IDLength+86])),
		FQDNResponderLength: uint16(binary.BigEndian.Uint16(packet.Payload[IDLength+86 : IDLength+88])),
		NATInitiatorIPv4:    [4]byte{packet.Payload[IDLength+88], packet.Payload[IDLength+89], packet.Payload[IDLength+90], packet.Payload[IDLength+91]},
		NATInitiatorIPv6: [16]byte{
			packet.Payload[IDLength+92], packet.Payload[IDLength+93], packet.Payload[IDLength+94], packet.Payload[IDLength+95],
			packet.Payload[IDLength+96], packet.Payload[IDLength+97], packet.Payload[IDLength+98], packet.Payload[IDLength+99],
			packet.Payload[IDLength+100], packet.Payload[IDLength+101], packet.Payload[IDLength+102], packet.Payload[IDLength+103],
			packet.Payload[IDLength+104], packet.Payload[IDLength+105], packet.Payload[IDLength+106], packet.Payload[IDLength+107],
		},
		NATResponderIPv4: [4]byte{packet.Payload[IDLength+108], packet.Payload[IDLength+109], packet.Payload[IDLength+110], packet.Payload[IDLength+111]},
		NATResponderIPv6: [16]byte{
			packet.Payload[IDLength+112], packet.Payload[IDLength+113], packet.Payload[IDLength+114], packet.Payload[IDLength+115],
			packet.Payload[IDLength+116], packet.Payload[IDLength+117], packet.Payload[IDLength+118], packet.Payload[IDLength+119],
			packet.Payload[IDLength+120], packet.Payload[IDLength+121], packet.Payload[IDLength+122], packet.Payload[IDLength+123],
			packet.Payload[IDLength+124], packet.Payload[IDLength+125], packet.Payload[IDLength+126], packet.Payload[IDLength+127],
		},
		InitiatorRealIPv4: [4]byte{packet.Payload[IDLength+128], packet.Payload[IDLength+129], packet.Payload[IDLength+130], packet.Payload[IDLength+131]},
		InitiatorRealIPv6: [16]byte{
			packet.Payload[IDLength+132], packet.Payload[IDLength+133], packet.Payload[IDLength+134], packet.Payload[IDLength+135],
			packet.Payload[IDLength+136], packet.Payload[IDLength+137], packet.Payload[IDLength+138], packet.Payload[IDLength+139],
			packet.Payload[IDLength+140], packet.Payload[IDLength+141], packet.Payload[IDLength+142], packet.Payload[IDLength+143],
			packet.Payload[IDLength+144], packet.Payload[IDLength+145], packet.Payload[IDLength+146], packet.Payload[IDLength+147],
		},
		ResponderRealIPv4: [4]byte{packet.Payload[IDLength+148], packet.Payload[IDLength+149], packet.Payload[IDLength+150], packet.Payload[IDLength+151]},
		ResponderRealIPv6: [16]byte{
			packet.Payload[IDLength+152], packet.Payload[IDLength+153], packet.Payload[IDLength+154], packet.Payload[IDLength+155],
			packet.Payload[IDLength+156], packet.Payload[IDLength+157], packet.Payload[IDLength+158], packet.Payload[IDLength+159],
			packet.Payload[IDLength+160], packet.Payload[IDLength+161], packet.Payload[IDLength+162], packet.Payload[IDLength+163],
			packet.Payload[IDLength+164], packet.Payload[IDLength+165], packet.Payload[IDLength+166], packet.Payload[IDLength+167],
		},
	}
	routeDirection.TunnelKey = packet.Payload[IDLength+168 : IDLength+168+int(routeDirection.TunnelKeyLength)]
	routeDirection.TemporaryKey = packet.Payload[IDLength+168+int(routeDirection.TunnelKeyLength) : IDLength+168+int(routeDirection.TunnelKeyLength)+int(routeDirection.TemporaryKeyLength)]
	routeDirection.InitiatorFQDN = packet.Payload[IDLength+168+int(routeDirection.TunnelKeyLength)+int(routeDirection.TemporaryKeyLength) : IDLength+168+int(routeDirection.TunnelKeyLength)+int(routeDirection.TemporaryKeyLength)+int(routeDirection.FQDNInitiatorLength)]
	routeDirection.ResponderFQDN = packet.Payload[IDLength+168+int(routeDirection.TunnelKeyLength)+int(routeDirection.TemporaryKeyLength)+int(routeDirection.FQDNInitiatorLength) : IDLength+168+int(routeDirection.TunnelKeyLength)+int(routeDirection.TemporaryKeyLength)+int(routeDirection.FQDNInitiatorLength)+int(routeDirection.FQDNResponderLength)]
	// Validate the lengths of TunnelKey, TemporaryKey, InitiatorFQDN, and ResponderFQDN
	if len(routeDirection.TunnelKey) != int(routeDirection.TunnelKeyLength) {
		return nil, fmt.Errorf("invalid tunnel key length: expected %d, got %d", routeDirection.TunnelKeyLength, len(routeDirection.TunnelKey))
	}
	if len(routeDirection.TemporaryKey) != int(routeDirection.TemporaryKeyLength) {
		return nil, fmt.Errorf("invalid temporary key length: expected %d, got %d", routeDirection.TemporaryKeyLength, len(routeDirection.TemporaryKey))
	}
	if len(routeDirection.InitiatorFQDN) != int(routeDirection.FQDNInitiatorLength) {
		return nil, fmt.Errorf("invalid initiator FQDN length: expected %d, got %d", routeDirection.FQDNInitiatorLength, len(routeDirection.InitiatorFQDN))
	}
	if len(routeDirection.ResponderFQDN) != int(routeDirection.FQDNResponderLength) {
		return nil, fmt.Errorf("invalid responder FQDN length: expected %d, got %d", routeDirection.FQDNResponderLength, len(routeDirection.ResponderFQDN))
	}
	// Check if the payload length matches the expected length
	expectedLength := BaseRouteDirectionLength +
		int(routeDirection.TunnelKeyLength) +
		int(routeDirection.TemporaryKeyLength) +
		int(routeDirection.FQDNInitiatorLength) +
		int(routeDirection.FQDNResponderLength) -
		BaseHeaderLength

	if len(packet.Payload) != expectedLength {
		return nil, fmt.Errorf("invalid packet length: expected %d, got %d", expectedLength, len(packet.Payload))
	}
	return routeDirection, nil
}

// ChangeRouteDirectionType changes route direction type.
func (r *RouteDirection) ChangeRouteDirectionType(typ TypeClass) {
	r.BaseHeader.Type = typ
}
