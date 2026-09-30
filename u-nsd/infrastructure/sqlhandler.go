//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE -self_package=github.com/Pluslab/cyphonic/nmsd/$GOPACKAGE
package infrastructure

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Pluslab/cyphonic/u-nsd/entity"
	"github.com/Pluslab/cyphonic/u-nsd/usecase/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var _ repository.SQLHandler = (*sqlHandler)(nil)

type sqlHandler struct {
	Conn                *gorm.DB
	notificationService *entity.NotificationService
}

func NewSQLHandler(host, username, databaseName, databasePort, sslMode, timezone, fqdn, nsIPv4, nsIPv6, strPort string) (*sqlHandler, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, username, databaseName, databasePort, sslMode, timezone)

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

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, fmt.Errorf("failed to convert port to integer: %v", err)
	}

	notificationService := entity.NotificationService{
		NSID:     fqdn,
		NSIPv4:   net.ParseIP(nsIPv4).To4(),
		NSIPv6:   net.ParseIP(nsIPv6).To16(),
		NSPort:   port,
		QuicFlag: false,
	}

	return &sqlHandler{
		Conn:                conn,
		notificationService: &notificationService,
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

func (s *sqlHandler) GetNodeInformationByNodeID(nodeID entity.ID) (*entity.NodeInformation, error) {
	var nodeInformation entity.NodeInformation
	if err := s.Conn.Where("node_id = ?", nodeID).First(&nodeInformation).Error; err != nil {
		return nil, fmt.Errorf("failed to get node information by node ID: %w", err)
	}

	return &nodeInformation, nil
}

func (s *sqlHandler) SaveInitiatorNodeAddress(interfaceName string, registrationRequest *entity.RegistrationRequest, natIPv4, natIPv6 net.IP, natPort int) (*entity.NodeAddress, error) {
	var nodeAddress entity.NodeAddress
	result := s.Conn.
		Limit(1).
		Where("node_id = ?", fmt.Sprintf("\\x%x", registrationRequest.BaseHeader.ID)).
		Find(&nodeAddress)

	realIPv4Bin := registrationRequest.NodeIPv4Address.As4()
	realIPv6Bin := registrationRequest.NodeIPv6Address.As16()

	nodeAddress = entity.NodeAddress{
		NodeID:        registrationRequest.BaseHeader.ID,
		InterfaceName: interfaceName,
		RealIPv4:      realIPv4Bin[:],
		RealIPv6:      realIPv6Bin[:],
		NATIPv4:       natIPv4.To4(),
		NATIPv6:       natIPv6.To16(),
		NATPort:       natPort,
		NSID:          s.notificationService.NSID,
		CreatedAt:     func(t time.Time) *time.Time { return &t }(time.Now()),
		UpdatedAt:     func(t time.Time) *time.Time { return &t }(time.Now()),
	}

	if result.RowsAffected == 0 {
		// Save Node Address to the database
		if err := s.Conn.Create(&nodeAddress).Error; err != nil {
			return nil, fmt.Errorf("failed to save node address: %w", err)
		}
	} else {
		// Update Node Address in the database
		if err := s.Conn.Model(&nodeAddress).Where("node_id = ?", fmt.Sprintf("\\x%x", registrationRequest.BaseHeader.ID)).Updates(nodeAddress).Error; err != nil {
			return nil, fmt.Errorf("failed to update node address: %w", err)
		}
	}

	return &nodeAddress, nil
}

func (s *sqlHandler) CreateSelfNotificationService() error {
	var notificationService entity.NotificationService

	result := s.Conn.
		Limit(1).
		Where("ns_id = ?", s.notificationService.NSID).
		Find(&notificationService)

	if result.RowsAffected != 0 {
		return fmt.Errorf("notification service with NSID %s already exists", s.notificationService.NSID)
	}

	// Save Node Address to the database
	if err := s.Conn.Create(s.notificationService).Error; err != nil {
		return fmt.Errorf("failed to set notification service: %w", err)
	}

	return nil
}

func (s *sqlHandler) DeleteNodeAddressesBySelfNSID() error {
	var notificationService entity.NotificationService
	var nodeAddress entity.NodeAddress

	result := s.Conn.
		Limit(1).
		Where("ns_id = ?", s.notificationService.NSID).
		Find(&notificationService)

	if result.RowsAffected == 0 {
		return nil
	}

	if err := s.Conn.Where("ns_id = ?", s.notificationService.NSID).Delete(&nodeAddress).Error; err != nil {
		return fmt.Errorf("failed to delete notification service: %w", err)
	}
	return nil

}

func (s *sqlHandler) DeleteSelfNotificationService() error {
	var notificationService entity.NotificationService

	result := s.Conn.
		Limit(1).
		Where("ns_id = ?", s.notificationService.NSID).
		Find(&notificationService)

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification service with NSID %s does not exist", s.notificationService.NSID)
	}

	if err := s.Conn.Where("ns_id = ?", s.notificationService.NSID).Delete(&notificationService).Error; err != nil { //s.notificationService
		return fmt.Errorf("failed to delete notification service: %w", err)
	}
	return nil

}
