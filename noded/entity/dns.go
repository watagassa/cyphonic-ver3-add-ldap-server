package entity

import (
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/google/gopacket"
	golayers "github.com/google/gopacket/layers"
	"github.com/miekg/dns"
)

// DNSQueue is the queue of dns.
type DNSQueue struct {
	TransactionID uint32
	FQDN          string
	DNSIdentifier DNSIdentifier
}

type DNSAnswerQueue struct {
	TransactionID uint32
	FQDN          string
	VirtualIPv4   net.IP
	VirtualIPv6   net.IP
	DNSIdentifier DNSIdentifier
}

// DNSIdentifier is the information required to identify a DNS query
type DNSIdentifier struct {
	ID   uint16
	Port uint16
}

// AnswerARecord creates A records.
// End Node gets destination node's network information from Node Management Service.
func AnswerARecord(dstPort, tid uint16, fqdn string, rnVirtualIPv4 net.IP) ([]byte, error) {
	msg := dns.Msg{
		MsgHdr: dns.MsgHdr{
			Response:           true,
			Opcode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Zero:               false,
			AuthenticatedData:  false,
			CheckingDisabled:   false,
			Rcode:              0,
		},
		Compress: false,
		Question: []dns.Question{},
		Answer:   []dns.RR{},
		Ns:       []dns.RR{},
		Extra:    []dns.RR{},
	}

	msg.MsgHdr.Id = uint16(tid)

	question := dns.Question{
		Name:   fqdn + ".",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}

	msg.Question = append(msg.Question, question)

	resp := new(dns.A)
	resp.Hdr = dns.RR_Header{
		Name:   fqdn + ".",
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    600,
	}

	resp.A = rnVirtualIPv4

	msg.Answer = append(msg.Answer, resp)

	payload, err := msg.Pack()
	if err != nil {
		config.LogErr("DNS packet packing failed", "error", err)
	}

	ip := golayers.IPv4{
		Version:  4,
		Protocol: golayers.IPProtocolUDP,
		SrcIP:    []byte{0xcb, 0x00, 0x71, 0x02},
		DstIP:    []byte{0xcb, 0x00, 0x71, 0x01},
	}

	udp := golayers.UDP{
		SrcPort: 53,
		DstPort: golayers.UDPPort(dstPort),
	}

	if err := udp.SetNetworkLayerForChecksum(&ip); err != nil {
		return nil, fmt.Errorf("udp set error: %w", err)
	}

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	buffer := gopacket.NewSerializeBuffer()

	if err := gopacket.SerializeLayers(buffer, options,
		&ip,
		&udp,
		gopacket.Payload(payload),
	); err != nil {
		return nil, fmt.Errorf("serialize error: %w", err)
	}

	outgoingPacket := buffer.Bytes()

	return outgoingPacket, nil
}

// AnswerAAAARecord creates AAAA records.
// End Node gets destination node's network information from Node Management Service.
func AnswerAAAARecord(dstPort, tid uint16, fqdn string, rnVirtualIPv6 net.IP) ([]byte, error) {
	msg := dns.Msg{
		MsgHdr: dns.MsgHdr{
			Response:           true,
			Opcode:             0,
			Authoritative:      false,
			Truncated:          false,
			RecursionDesired:   true,
			RecursionAvailable: true,
			Zero:               false,
			AuthenticatedData:  false,
			CheckingDisabled:   false,
			Rcode:              0,
		},
		Compress: false,
		Question: []dns.Question{},
		Answer:   []dns.RR{},
		Ns:       []dns.RR{},
		Extra:    []dns.RR{},
	}

	msg.MsgHdr.Id = uint16(tid)

	question := dns.Question{
		Name:   fqdn + ".",
		Qtype:  dns.TypeAAAA,
		Qclass: dns.ClassINET,
	}

	msg.Question = append(msg.Question, question)

	resp := new(dns.AAAA)
	resp.Hdr = dns.RR_Header{
		Name:   fqdn + ".",
		Rrtype: dns.TypeAAAA,
		Class:  dns.ClassINET,
		Ttl:    600,
	}

	resp.AAAA = rnVirtualIPv6

	msg.Answer = append(msg.Answer, resp)

	payload, err := msg.Pack()
	if err != nil {
		config.LogErr("DNS packet packing failed", "error", err)
	}

	ip := golayers.IPv6{
		Version:      6,
		TrafficClass: 0,
		FlowLabel:    0,
		Length:       0,
		NextHeader:   golayers.IPProtocolUDP,
		HopLimit:     64,
		SrcIP:        net.ParseIP("2003::8"),
		DstIP:        net.ParseIP("2003::88"),
	}

	udp := golayers.UDP{
		SrcPort: 53,
		DstPort: golayers.UDPPort(dstPort),
	}

	if err := udp.SetNetworkLayerForChecksum(&ip); err != nil {
		return nil, fmt.Errorf("udp set error: %w", err)
	}

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	buffer := gopacket.NewSerializeBuffer()

	if err := gopacket.SerializeLayers(buffer, options,
		&ip,
		&udp,
		gopacket.Payload(payload),
	); err != nil {
		return nil, fmt.Errorf("serialize error: %w", err)
	}

	outgoingPacket := buffer.Bytes()

	return outgoingPacket, nil
}
