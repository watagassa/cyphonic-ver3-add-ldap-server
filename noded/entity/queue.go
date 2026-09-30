package entity

import (
	"net"
	"net/netip"
	"sync"
)

type InPacketQueue struct {
	Buffer []byte
	Addr   *net.UDPAddr
}

type ParsedPacketQueue struct {
	Packet *BasePacket
	Addr   *net.UDPAddr
}

type OutPacketQueue struct {
	Buffer []byte
	Addr   *net.UDPAddr
	Soc    Socket
}

type CapsuleMsgQueue struct {
	sync.Mutex
	Buffer []byte
	Addr   *net.UDPAddr
	Soc    Socket
}

type OutboundMsgQueue struct {
	Data []byte
	Dst  *netip.Addr
}

type StagedOutboundMsgQueue struct {
	Data     []byte
	Dst      *netip.Addr
	QueuePtr *CapsuleMsgQueue
}

type EncryptedMsgQueue struct {
	Packet *BasePacket
	EndKey []byte
}

type StagedInboundMsgQueue struct {
	Packet   *BasePacket
	QueuePtr *InboundMsgQueue
}

type InboundMsgQueue struct {
	sync.Mutex
	Data []byte
}
