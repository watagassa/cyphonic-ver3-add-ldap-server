package infrastructure

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Pluslab/cyphonic/dsd/entity"
	"github.com/Pluslab/cyphonic/dsd/usecase/repository"
	"github.com/redis/go-redis/v9"
)

var ErrAssertInitiatorNodeID = errors.New("failed to assert type for initiatorNodeID")

var (
	_ repository.RedisClient = (*redisClient)(nil)
	_ repository.RedisClient = (*redisClusterClient)(nil)
)

type redisClient struct {
	client *redis.Client
}

type redisClusterClient struct {
	client *redis.ClusterClient
}

func NewRedisClient(addrs []string) (*redisClient, error) {
	return &redisClient{
		client: redis.NewClient(
			&redis.Options{
				Addr: addrs[0],
			}),
	}, nil
}

func NewRedisClusterClient(addrs []string) (*redisClusterClient, error) {
	return &redisClusterClient{
		client: redis.NewClusterClient(
			&redis.ClusterOptions{
				Addrs: addrs,
			}),
	}, nil
}

func (r *redisClient) Close() error {
	return fmt.Errorf("failed to close redis client: %w", r.client.Close())
}

func (r *redisClusterClient) Close() error {
	return fmt.Errorf("failed to close redis cluster client: %w", r.client.Close())
}

func createOrUpdatePathInformation(ctx context.Context, client redis.Cmdable, pathID *entity.ID,
	initiatorNodeAddress, responderNodeAddress *entity.NodeAddress, tunnelKey []byte,
) error {
	pathInformation := &entity.PathInformation{
		GeneratingFlag:      true,
		InitiatorIPv4:       initiatorNodeAddress.NATIPv4,
		InitiatorIPv6:       initiatorNodeAddress.NATIPv6,
		InitiatorPort:       uint16(initiatorNodeAddress.NATPort),
		ResponderIPv4:       responderNodeAddress.NATIPv4,
		ResponderIPv6:       responderNodeAddress.NATIPv6,
		ResponderPort:       uint16(responderNodeAddress.NATPort),
		TunnelKey:           tunnelKey,
		TunnelKeyCipherType: uint16(entity.AES256CBC),
		TunnelKeyLength:     uint16(len(tunnelKey)),
	}
	pathInformationField := entity.GetPathInformationField()
	sliceCmd := client.HMGet(ctx, fmt.Sprintf("%x", pathID), pathInformationField...)

	value, err := sliceCmd.Result()
	if err != nil {
		return fmt.Errorf("failed to get hash map: %w", err)
	}

	if value[0] != nil {
		pathInformation = entity.GeneratePathInformation(value)
		pathInformation.GeneratingFlag = true
		pathInformation.TunnelKey = []byte(hex.EncodeToString(tunnelKey))
		pathInformation.TunnelKeyLength = uint16(len(tunnelKey))
		pathInformation.TunnelKeyCipherType = uint16(entity.AES256CBC)
	}

	pathInformationInterface := map[string]interface{}{
		"generatingFlag":      pathInformation.GeneratingFlag,
		"initiatorIPv4":       pathInformation.InitiatorIPv4,
		"initiatorIPv6":       pathInformation.InitiatorIPv6,
		"initiatorPort":       pathInformation.InitiatorPort,
		"responderIPv4":       pathInformation.ResponderIPv4,
		"responderIPv6":       pathInformation.ResponderIPv6,
		"responderPort":       pathInformation.ResponderPort,
		"tunnelKey":           pathInformation.TunnelKey,
		"tunnelKeyCipherType": pathInformation.TunnelKeyCipherType,
		"tunnelKeyLength":     pathInformation.TunnelKeyLength,
	}

	boolCmd := client.HMSet(ctx, fmt.Sprintf("%x", pathID), pathInformationInterface)
	if err := boolCmd.Err(); err != nil {
		return fmt.Errorf("failed to save or update path information: %w", err)
	}

	return nil
}

func (r *redisClient) CreateOrUpdatePathInformation(ctx context.Context, pathID *entity.ID,
	initiatorNodeAddress, responderNodeAddress *entity.NodeAddress, tunnelKey []byte,
) error {
	return createOrUpdatePathInformation(ctx, r.client, pathID, initiatorNodeAddress, responderNodeAddress, tunnelKey)
}

func (r *redisClusterClient) CreateOrUpdatePathInformation(ctx context.Context, pathID *entity.ID,
	initiatorNodeAddress, responderNodeAddress *entity.NodeAddress, tunnelKey []byte,
) error {
	return createOrUpdatePathInformation(ctx, r.client, pathID, initiatorNodeAddress, responderNodeAddress, tunnelKey)
}

func (r *redisClient) GetInitiatorNodeIDByPathID(ctx context.Context, pathID *entity.ID) (entity.ID, error) {
	sliceCmd := r.client.HMGet(ctx, fmt.Sprintf("%x", pathID), "initiatorNodeID")

	value, err := sliceCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash map: %w", err)
	}

	if value[0] == nil {
		return nil, nil
	}

	initiatorNodeIDStr, ok := value[0].(string)
	if !ok {
		return nil, ErrAssertInitiatorNodeID
	}

	return entity.ID(initiatorNodeIDStr), nil
}

func (r *redisClusterClient) GetInitiatorNodeIDByPathID(ctx context.Context, pathID *entity.ID) (entity.ID, error) {
	sliceCmd := r.client.HMGet(ctx, fmt.Sprintf("%x", pathID), "initiatorNodeID")

	value, err := sliceCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash map: %w", err)
	}

	if value[0] == nil {
		return nil, nil
	}

	initiatorNodeIDStr, ok := value[0].(string)
	if !ok {
		return nil, ErrAssertInitiatorNodeID
	}

	return entity.ID(initiatorNodeIDStr), nil
}

func (r *redisClient) SetInitiatorNodeIDAndResponderFQDN(ctx context.Context, pathID *entity.ID, initiatorNodeID entity.ID, responderFQDN string) error {
	pathInformationInterface := map[string]interface{}{
		"initiatorNodeID": initiatorNodeID,
		"responderFQDN":   responderFQDN,
	}

	boolCmd := r.client.HMSet(ctx, fmt.Sprintf("%x", pathID), pathInformationInterface)
	if err := boolCmd.Err(); err != nil {
		return fmt.Errorf("failed to set initiator node ID and responder FQDN: %w", err)
	}

	return nil
}

func (r *redisClusterClient) SetInitiatorNodeIDAndResponderFQDN(ctx context.Context, pathID *entity.ID, initiatorNodeID entity.ID, responderFQDN string) error {
	pathInformationInterface := map[string]interface{}{
		"initiatorNodeID": initiatorNodeID,
		"responderFQDN":   responderFQDN,
	}

	boolCmd := r.client.HMSet(ctx, fmt.Sprintf("%x", pathID), pathInformationInterface)
	if err := boolCmd.Err(); err != nil {
		return fmt.Errorf("failed to set initiator node ID and responder FQDN: %w", err)
	}

	return nil
}
