package infrastructure

import (
	"fmt"
	"net"
	"time"

	"github.com/Pluslab/cyphonic/dsd/entity"
	"github.com/Pluslab/cyphonic/dsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/dsd/usecase/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ repository.SQLHandler = (*sqlHandler)(nil)

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

func (s *sqlHandler) Close() error {
	sqlDB, err := s.Conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}

func (s *sqlHandler) GetNodeInformationAndNodeAddressByNodeID(nodeID string) (entity.NodeInformation, entity.NodeAddress, error) {
	var nodeInformation entity.NodeInformation
	var nodeAddress entity.NodeAddress

	err := s.Conn.Where("node_id = ?", nodeID).First(&nodeInformation).Error
	if err != nil {
		return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node information: %w", err)
	}

	err = s.Conn.Where("node_id = ?", nodeID).First(&nodeAddress).Error
	if err != nil {
		return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node address: %w", err)
	}

	return nodeInformation, nodeAddress, nil
}

func (s *sqlHandler) GetNodeInformationAndNodeAddressByFQDN(fqdn string) (entity.NodeInformation, entity.NodeAddress, error) {
	var nodeAddress entity.NodeAddress
	var nodeInformation entity.NodeInformation

	err := s.Conn.Where("fqdn = ?", fqdn).First(&nodeInformation).Error
	if err != nil {
		return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node information: %w", err)
	}

	err = s.Conn.Where("node_id = ?", nodeInformation.NodeID).First(&nodeAddress).Error
	if err != nil {
		return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node address: %w", err)
	}

	//err := s.Conn.Where("address_value = ?", fqdn).First(&nodeAddress).Error
	//if err != nil {
	//	return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node address: %w", err)
	//}
	//
	//err = s.Conn.Where("node_id = ?", nodeAddress.NodeID).First(&nodeInformation).Error
	//if err != nil {
	//	return entity.NodeInformation{}, entity.NodeAddress{}, fmt.Errorf("failed to get node information: %w", err)
	//}

	return nodeInformation, nodeAddress, nil
}

func (s *sqlHandler) GetNotificationServiceByNSID(nsid string) (entity.NotificationService, error) {
	var notificationService entity.NotificationService

	err := s.Conn.Where("ns_id = ?", nsid).First(&notificationService).Error
	if err != nil {
		return entity.NotificationService{}, fmt.Errorf("failed to get notification service: %w", err)
	}

	return notificationService, nil
}

func (s *sqlHandler) GetRandomUDPTunnelRelayService() (entity.TunnelRelayService, error) {
	var cnt int64

	rowsCnt := s.Conn.Table("tunnel_relay_services").Count(&cnt)
	if rowsCnt.Error != nil {
		return entity.TunnelRelayService{}, fmt.Errorf("failed to count tunnel relay services: %w", rowsCnt.Error)
	}

	if cnt == 0 {
		return entity.TunnelRelayService{}, fmt.Errorf("no tunnel relay services available")
	}

	var tunnelRelayService entity.TunnelRelayService
	err := s.Conn.Where("quic_flag = ?", false).Order("RANDOM()").First(&tunnelRelayService).Error
	if err != nil {
		return entity.TunnelRelayService{}, fmt.Errorf("failed to get random tunnel relay service: %w", err)
	}

	return tunnelRelayService, nil
}

// In the future, I'd like to change the processing to use Redis.
func (s *sqlHandler) SetPathInformation(pathID, tunnelKey []byte,
	initiatorNodeAddress, responderNodeAddress entity.NodeAddress,
) error {
	pathInfo := entity.PathInformation{
		PathID:                []byte{},
		GeneratingFlag:        false,
		InitiatorIPv4:         net.IP{},
		InitiatorIPv6:         net.IP{},
		InitiatorPort:         0,
		InitiatorConnectionID: []byte{},
		ResponderIPv4:         net.IP{},
		ResponderIPv6:         net.IP{},
		ResponderPort:         0,
		ResponderConnectionID: []byte{},
		TunnelKey:             []byte{},
		TunnelKeyCipherType:   0,
		TunnelKeyLength:       0,
		TunnelKeyExpire:       time.Time{},
		CreatedAt:             time.Time{},
		UpdatedAt:             time.Time{},
	}
	result := s.Conn.Limit(1).
		Where("path_id = ?", fmt.Sprintf("\\x%x", pathID)).
		Find(&pathInfo)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// insert path information
		pathInfo = entity.PathInformation{
			PathID:                pathID,
			GeneratingFlag:        false,
			InitiatorIPv4:         initiatorNodeAddress.RealIPv4,
			InitiatorIPv6:         initiatorNodeAddress.RealIPv6,
			InitiatorPort:         30000,
			InitiatorConnectionID: []byte{}, // FIXME
			ResponderIPv4:         responderNodeAddress.RealIPv4,
			ResponderIPv6:         responderNodeAddress.RealIPv6,
			ResponderPort:         30000,
			ResponderConnectionID: []byte{}, // FIXME
			TunnelKey:             tunnelKey,
			TunnelKeyCipherType:   0,
			TunnelKeyLength:       uint16(len(tunnelKey)),
			TunnelKeyExpire:       time.Now().AddDate(0, 0, 30), // FIXME: 仮
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		config.LogDebug("Set path information", "path_id", pathID)

		if err := s.Conn.Create(&pathInfo).Error; err != nil {
			return err
		}
	} else {
		// update node information
		updatePathInfo := entity.PathInformation{
			PathID:                pathID,
			GeneratingFlag:        false,
			InitiatorIPv4:         initiatorNodeAddress.NATIPv4,
			InitiatorIPv6:         initiatorNodeAddress.NATIPv6,
			InitiatorPort:         uint16(initiatorNodeAddress.NATPort),
			InitiatorConnectionID: []byte{}, // FIXME
			ResponderIPv4:         responderNodeAddress.NATIPv4,
			ResponderIPv6:         responderNodeAddress.NATIPv6,
			ResponderPort:         uint16(responderNodeAddress.NATPort),
			ResponderConnectionID: []byte{}, // FIXME
			TunnelKey:             tunnelKey,
			TunnelKeyCipherType:   0,
			TunnelKeyLength:       uint16(len(tunnelKey)),
			TunnelKeyExpire:       time.Now().AddDate(0, 0, 30), // FIXME: 仮
			CreatedAt:             pathInfo.CreatedAt,
			UpdatedAt:             time.Now(),
		}

		config.LogDebug("Updated path information", "path_id", pathID)

		if err := s.Conn.Model(&pathInfo).Where("path_id = ?", fmt.Sprintf("\\x%x", pathID)).Updates(updatePathInfo).Error; err != nil {
			return err
		}
	}

	return nil
}
