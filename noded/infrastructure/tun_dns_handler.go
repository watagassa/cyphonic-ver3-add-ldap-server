package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
	"github.com/Pluslab/cyphonic/noded/usecase/repository"
	"github.com/google/gopacket"
	golayers "github.com/google/gopacket/layers"
	"github.com/miekg/dns"
)

var _ repository.TunDnsHandler = (*tunDnsHandler)(nil)

// TODO: ip.goの作成
const (
	VIPv4DNSTunPrefix = 24
	VIPv6DNSTunPrefix = 64
)

type tunDnsHandler struct {
	isVirtualIPv6       bool
	tunDnsInterfaceName string
	tunDnsInterfaceIPv4 string
	tunDnsInterfaceIPv6 string
	transactionID       uint32
	tunInterface        *entity.TunInterface
}

func NewTunDnsHandler(virtualIPType int, tunDnsInterfaceName, tunDnsInterfaceIPv4, tunDnsInterfaceIPv6 string, transactionID uint32) (repository.TunDnsHandler, error) {
	var (
		isVirtualIPv6 bool
		tun           *entity.TunInterface
		err           error
	)

	switch virtualIPType {
	case 4:
		tun, err = entity.NewTunInterface(tunDnsInterfaceName, tunDnsInterfaceIPv4, VIPv4DNSTunPrefix)
		isVirtualIPv6 = false
	case 6:
		tun, err = entity.NewTunInterface(tunDnsInterfaceName, tunDnsInterfaceIPv6, VIPv6DNSTunPrefix)
		isVirtualIPv6 = true
	default:
		return nil, fmt.Errorf("unknown ip version: %d", virtualIPType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create tun interface: %w", err)
	}

	if err = tun.Up(); err != nil {
		_ = tun.Close()
		return nil, fmt.Errorf("failed to up tun interface: %w", err)
	}

	config.LogDebug("Start virtual DNS I/F",
		"tun_name", tun.Name(),
		"vip_addr", tun.Address())

	return &tunDnsHandler{
		isVirtualIPv6:       isVirtualIPv6,
		tunDnsInterfaceName: tunDnsInterfaceName,
		tunInterface:        tun,
		tunDnsInterfaceIPv4: tunDnsInterfaceIPv4,
		tunDnsInterfaceIPv6: tunDnsInterfaceIPv6,
		transactionID:       transactionID,
	}, nil
}

func (t *tunDnsHandler) Read(ctx context.Context, dnsQueue chan<- entity.DNSQueue) error {
	config.LogDebug("Starting to read data from TUN DNS interface")
	defer config.LogDebug("Finished reading data from TUN DNS interface")

	stop := context.AfterFunc(ctx, func() {
		_ = t.tunInterface.Close()
	})
	defer stop()

	for {
		buf := make([]byte, 1500)
		size, err := t.tunInterface.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
				config.LogDebug("TUN_DNS interface closed, stopping data reception")
				return nil
			}
			config.LogErr("Failed to get DNS packet", "error", err)
		}

		var packet gopacket.Packet

		// TODO: Atomic boolを使用する
		// TODO: t.isVirtualIPv6を後から変更するか確認する

		// convert the buffer to a packet format
		switch t.isVirtualIPv6 {
		case false:
			packet = gopacket.NewPacket(buf[:size], golayers.LayerTypeIPv4, gopacket.Default)
			if err != nil {
				config.LogErr("Failed to create Tunnel Interface Virtual IPv4", "error", err)
				continue
			}
		case true:
			packet = gopacket.NewPacket(buf[:size], golayers.LayerTypeIPv6, gopacket.Default)
			if err != nil {
				config.LogErr("Failed to create Tunnel Interface Virtual IPv6", "error", err)
				continue
			}
		}

		// extract the UDP layer from the packet
		udpLayer := packet.Layer(golayers.LayerTypeUDP)

		if udpLayer == nil {
			continue
		}

		// type conversion
		udp, ok := udpLayer.(*golayers.UDP)
		if !ok {
			config.LogErr("Failed to unpack udp packet", "error", ok)
			continue
		}

		// put the analysis results of the UDP header and payload in msg
		msg := new(dns.Msg)
		if err = msg.Unpack(udp.Payload); err != nil {
			config.LogErr("dns: Failed to unpack packet", "error", err)
			continue
		}

		for _, v := range msg.Question {
			if !strings.Contains(v.Name, "cyphonic.org") {
				continue
			}

			switch t.isVirtualIPv6 {
			case false:
				if v.Qtype != dns.TypeA {
					continue
				}

				config.LogDebug("dns: Received v4 DNS Query", "transaction_id", msg.MsgHdr.Id, "question", v)

				domainName := strings.TrimRight(v.Name, ".")

				dnsQueue <- entity.DNSQueue{
					TransactionID: t.transactionID,
					FQDN:          domainName,
					DNSIdentifier: entity.DNSIdentifier{
						ID:   msg.MsgHdr.Id,
						Port: uint16(udp.SrcPort),
					},
				}

			case true:
				if v.Qtype != dns.TypeAAAA {
					continue
				}

				config.LogDebug("dns: Received v6 DNS Query", "transaction_id", msg.MsgHdr.Id, "question", v)

				domainName := strings.TrimRight(v.Name, ".")

				dnsQueue <- entity.DNSQueue{
					TransactionID: t.transactionID,
					FQDN:          domainName,
					DNSIdentifier: entity.DNSIdentifier{
						ID:   msg.MsgHdr.Id,
						Port: uint16(udp.SrcPort),
					},
				}
			}
		}
	}
}

func (t *tunDnsHandler) Write(ctx context.Context,
	dnsAnswerQueue <-chan entity.DNSAnswerQueue,
	cachedDnsAnswerQueue <-chan entity.DNSAnswerQueue) error {
	config.LogDebug("Starting to write data to TUN DNS interface")
	defer config.LogDebug("Finished writing data to TUN DNS interface")

	stop := context.AfterFunc(ctx, func() {
		_ = t.tunInterface.Close()
	})
	defer stop()

	for {
		var ans entity.DNSAnswerQueue

		select {
		case <-ctx.Done():
			return nil
		case ans = <-dnsAnswerQueue:
		case ans = <-cachedDnsAnswerQueue:
		}

		// TODO: 32bitに統一
		var dnsPacket []byte
		var err error
		if !t.isVirtualIPv6 {
			dnsPacket, err = entity.AnswerARecord(ans.DNSIdentifier.Port, ans.DNSIdentifier.ID, ans.FQDN, ans.VirtualIPv4)
		} else {
			dnsPacket, err = entity.AnswerAAAARecord(ans.DNSIdentifier.Port, ans.DNSIdentifier.ID, ans.FQDN, ans.VirtualIPv6)
		}

		if err != nil {
			config.LogErr("Failed to create DNS answer packet", "error", err)
			continue
		}

		_, err = t.tunInterface.Write(dnsPacket)
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
				config.LogDebug("TUN_DNS interface closed, stopping data transmission")
				return nil
			}
			config.LogErr("Failed to write DNS packet", "error", err)
			continue
		}

		config.LogDebug("Wrote DNS answer packet as A record or AAAA record",
			"fqdn", ans.FQDN,
			"virtual_ipv4", ans.VirtualIPv4.String(),
			"virtual_ipv6", ans.VirtualIPv6.String(),
		)
	}
}
