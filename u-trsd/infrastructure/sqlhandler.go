package infrastructure

import (
	"bytes"
	"fmt"
	"net"

	"github.com/Pluslab/cyphonic/u-trsd/entity"
	"github.com/Pluslab/cyphonic/u-trsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/u-trsd/usecase/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ repository.SQLHandler = (*sqlHandler)(nil)

type sqlHandler struct {
	Conn *gorm.DB
	FQDN string
}

// NewSQLHandler creates a new SQL handler with the provided database connection parameters.
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

// CreateTunnelRelayService creates a new tunnel relay service in the database.
// This is intended for self-registration of the tunnel relay service.
func (s *sqlHandler) CreateTunnelRelayService(trs entity.TunnelRelayService) error {
	result := s.Conn.
		Limit(1).
		Where("trs_id = ?", trs.TunnleRelayServiceID).
		Find(&trs)

	if result.RowsAffected != 0 {
		return fmt.Errorf("tunnel relay service with ID %s already exists", trs.TunnleRelayServiceID)
	}

	// Save Node Address to the database
	if err := s.Conn.Create(trs).Error; err != nil {
		return fmt.Errorf("failed to set tunnel relay service: %w", err)
	}

	return nil
}

// CreateTunnelRelayService creates a new tunnel relay service in the database.
// This is intended for self-registration of the tunnel relay service.
func (s *sqlHandler) DeleteTunnelRelayService(trsID string) error {
	var trs entity.TunnelRelayService

	result := s.Conn.
		Limit(1).
		Where("trs_id = ?", trsID).
		Find(&trs)

	if result.RowsAffected == 0 {
		return fmt.Errorf("tunnel relay service with ID %s does not exist", trsID)
	}

	if err := s.Conn.Where("trs_id = ?", trsID).Delete(&trs).Error; err != nil {
		return fmt.Errorf("failed to delete tunnel relay service: %w", err)
	}
	return nil
}

// GetPathInformationByPathID retrieves path information by path ID.
func (s *sqlHandler) GetPathInformationByPathID(pathID string) (entity.PathInformation, error) {
	// convert pathID to bytes
	pathIDBytes := []byte(pathID)

	// trim the bytes
	trimmedBytes := bytes.TrimRight(pathIDBytes, "\x00")

	var pathInformation entity.PathInformation

	err := s.Conn.Where("path_id = ?", trimmedBytes).First(&pathInformation).Error
	if err != nil {
		return entity.PathInformation{}, fmt.Errorf("failed to get path information: %w", err)
	}

	return pathInformation, nil
}

// SetResponderNodeAddress updates the responder node address for a given path ID.
func (s *sqlHandler) SetResponderNodeAddress(pathID string, address net.IP, port uint16) error {
	// Convert to IPv4
	ip4 := address.To4()

	// If IPv6, ip4 will be nil.
	if ip4 == nil {
		return s.setResponderNodeAddress(pathID, net.IPv4zero, address, port)
	}
	return s.setResponderNodeAddress(pathID, ip4, net.IPv6zero, port)
}

func (s *sqlHandler) setResponderNodeAddress(pathID string, ipv4, ipv6 net.IP, port uint16) error {
	// convert pathID to bytes
	pathIDBytes := []byte(pathID)

	// trim the bytes
	trimmedBytes := bytes.TrimRight(pathIDBytes, "\x00")

	// Set RN to responder_port and responder_ipv4/responder_ipv6 on path_informations
	err := s.Conn.Model(&entity.PathInformation{}).Where("path_id = ?", trimmedBytes).
		Updates(map[string]interface{}{
			"responder_port": port,
			"responder_ipv4": ipv4,
			"responder_ipv6": ipv6,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update responder address: %w", err)
	}

	config.LogDebug("Updated responder node address on database", "pathid", pathID, "ipv4", ipv4, "ipv6", ipv6, "port", port)

	return nil
}

// SetInitiatorNodeAddress updates the initiator node address for a given path ID.
func (s *sqlHandler) SetInitiatorNodeAddress(pathID string, address net.IP, port uint16) error {
	// Convert to IPv4
	ip4 := address.To4()

	// If IPv6, ip4 will be nil.
	if ip4 == nil {
		return s.setInitiatorNodeAddress(pathID, net.IPv4zero, address, port)
	}
	return s.setInitiatorNodeAddress(pathID, ip4, net.IPv6zero, port)
}

func (s *sqlHandler) setInitiatorNodeAddress(pathID string, ipv4, ipv6 net.IP, port uint16) error {
	// convert pathID to bytes
	pathIDBytes := []byte(pathID)

	// trim the bytes
	trimmedBytes := bytes.TrimRight(pathIDBytes, "\x00")

	// Set IN to initiator_port and initiator_ipv4/initiator_ipv6 on path_informations
	err := s.Conn.Model(&entity.PathInformation{}).Where("path_id = ?", trimmedBytes).
		Updates(map[string]interface{}{
			"initiator_port": port,
			"initiator_ipv4": ipv4,
			"initiator_ipv6": ipv6,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update initiator address: %w", err)
	}

	config.LogDebug("Updated initiator node address on database", "pathid", pathID, "ipv4", ipv4, "ipv6", ipv6, "port", port)

	return nil
}
