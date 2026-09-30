package usecase

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Pluslab/cyphonic/noded/entity"
	"github.com/Pluslab/cyphonic/noded/infrastructure/config"
)

func (h *handler) HandleDNSQueries(ctx context.Context,
	dnsQueue <-chan entity.DNSQueue,
	cachedDnsAnswerQueue chan<- entity.DNSAnswerQueue,
	directionRequestQueue chan<- entity.OutPacketQueue) error {
	config.LogDebug("Starting handle dns queries function")
	defer config.LogDebug("Finished handle dns queries function")

	for {
		var msg entity.DNSQueue

		select {
		case <-ctx.Done():
			return nil
		case msg = <-dnsQueue:
		}

		// if the peer already exists, write the DNS Answer
		if peer, ok := h.peerRegistry.GetPeerByFQDN(msg.FQDN); ok {
			config.LogDebug("Peer already exists in cache, skipping Direction Request")

			cachedDnsAnswerQueue <- entity.DNSAnswerQueue{
				TransactionID: peer.TransactionID(),
				FQDN:          string(peer.FQDN()),
				VirtualIPv4:   peer.VirtualIPv4().AsSlice(),
				VirtualIPv6:   peer.VirtualIPv6().AsSlice(),
				DNSIdentifier: msg.DNSIdentifier,
			}
			continue
		}

		buf, sendAddress, err := h.generateDirectionRequest([]byte(msg.FQDN))

		if err != nil {
			config.LogErr("Failed to generate Direction Request", "error", err)
			continue
		}

		if buf == nil || sendAddress == nil {
			continue
		}

		// cache DNS information
		h.dnsCache.Put(msg.TransactionID, msg.DNSIdentifier.Port, msg.DNSIdentifier.ID, time.Now().Add(10*time.Second).UnixNano())

		// if the peer is new, send a Direction Request
		directionRequestQueue <- entity.OutPacketQueue{
			Buffer: buf,
			Addr:   sendAddress,
		}
	}
}

func (h *handler) generateDirectionRequest(rnFQDN []byte) ([]byte, *net.UDPAddr, error) {
	directionRequest, err := entity.GenerateDirectionRequest(h.baseHeader, h.fqdn, rnFQDN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate direction request: %w", err)
	}

	config.LogDebug("Generated Direction Request",
		"transaction_id", directionRequest.BaseHeader.TransactionID,
		"path_id", directionRequest.PathID,
		"direction_request", directionRequest)

	buffer, err := directionRequest.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal direction request: %w", err)
	}

	// TODO: 送信前にcommonKeyで暗号化しないといけない
	sendAddress, err := h.nsAddress.UDPAddr(h.localIPVersion)
	if err != nil {
		config.LogErr("Failed to generate UDP Address from local IP", "error", err)
		return nil, nil, fmt.Errorf("failed to generate UDP address: %w", err)
	}

	config.LogDebug("Sent Direction Request to NS", "dst", sendAddress.IP)

	return buffer, &sendAddress, nil
}
