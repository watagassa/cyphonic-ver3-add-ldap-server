package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Pluslab/cyphonic/psd/entity"
	"github.com/Pluslab/cyphonic/psd/infrastructure/config"
	"github.com/Pluslab/cyphonic/psd/usecase/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type sqlHandler struct {
	Conn *gorm.DB
	FQDN string
}

func NewSQLHandler(host, user, name, port, sslMode, timeZone, fqdn string) (repository.SQLHandler, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, name, port, sslMode, timeZone)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction:                   false,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		Logger:                                   nil,
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		IgnoreRelationshipsWhenMigrating:         false,
		DisableNestedTransaction:                 false,
		AllowGlobalUpdate:                        false,
		QueryFields:                              false,
		CreateBatchSize:                          0,
		TranslateError:                           false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	})
	if err != nil {
		return nil, fmt.Errorf("can't open gorm: %w", err)
	}

	return &sqlHandler{
		Conn: conn,
		FQDN: fqdn,
	}, nil
}

var ErrAliasFQDNNotFound error = errors.New("alias fqdn not found")

func (s *sqlHandler) GetNodeInformation(ctx context.Context, nodeID entity.ID) (entity.NodeInformation, error) {
	nodeInfo := entity.NodeInformation{
		Version:            "",
		FQDN:               "",
		DeviceID:           "",
		NodeID:             []byte{},
		VirtualIPv4:        []byte{},
		VirtualIPv4Netmask: []byte{},
		VirtualIPv6:        []byte{},
		VirtualIPv6Prefix:  []byte{},
		ApplicationID:      "",
		ApplicationPort:    0,
		NotificationType:   0,
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
	}

	result := s.Conn.Limit(1).Where("node_id = ?", fmt.Sprintf("\\x%x", nodeID)).Find(&nodeInfo)
	if result.Error != nil || result.RowsAffected == 0 {
		return nodeInfo, fmt.Errorf("can't get node information: %w", result.Error)
	}

	return nodeInfo, nil
}

func (s *sqlHandler) GetAliasFQDN(ctx context.Context, deviceID, desiredFQDN string) (entity.FQDNAlias, error) {
	fqdnAlias := entity.FQDNAlias{
		DeviceID:  "",
		FQDN:      "",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	result := s.Conn.Limit(1).Where("device_id = ? AND fqdn_alias = ?", deviceID, desiredFQDN).Find(&fqdnAlias)
	if result.Error != nil {
		return fqdnAlias, fmt.Errorf("failed to get alias fqdn: %w", result.Error)
	} else if result.RowsAffected == 0 {
		return fqdnAlias, ErrAliasFQDNNotFound
	}

	return fqdnAlias, nil
}

func (s *sqlHandler) UpdateNodeInformation(ctx context.Context, nodeInfo entity.NodeInformation) error {
	updateNodeInformation := nodeInfo
	updateNodeInformation.UpdatedAt = time.Now()

	if err := s.Conn.Model(&nodeInfo).Where("node_id = ?", fmt.Sprintf("\\x%x", nodeInfo.NodeID)).Updates(updateNodeInformation).Error; err != nil {
		return fmt.Errorf("failed to update node information: %w", err)
	}

	config.LogDebug("Updated node information", "information", updateNodeInformation)

	return nil
}
