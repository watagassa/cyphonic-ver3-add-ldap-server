package infrastructure

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"time"

	"github.com/Pluslab/cyphonic/cmsd/entity"
	"github.com/Pluslab/cyphonic/cmsd/infrastructure/config"
	"github.com/Pluslab/cyphonic/cmsd/usecase/repository"
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

func (s *sqlHandler) GetRandomNotificationService() (*entity.NotificationServiceInfomation, entity.StatusClass, error) {
	var cnt int64

	NSInfoErr := entity.NotificationServiceInfomation{
		ID:   "",
		IPv4: netip.AddrFrom4([4]byte(net.IPv4zero.To4())),
		IPv6: netip.AddrFrom16([16]byte(net.IPv6zero.To16())),
		Port: 0,
	}

	notificationService := entity.NotificationService{
		NSid:     "",
		NSIPv4:   []byte{},
		NSIPv6:   []byte{},
		NSPort:   0,
		QuicFlag: false,
	}

	rowsCnt := s.Conn.Table("notification_services").Count(&cnt)
	if rowsCnt.Error != nil {
		return &NSInfoErr, entity.StatusClassInternalServerError, rowsCnt.Error
	}

	randNumber, err := rand.Int(rand.Reader, big.NewInt(cnt))
	if err != nil {
		return &NSInfoErr, entity.StatusClassInternalServerError, err
	}

	result := s.Conn.Limit(1).Offset(int(randNumber.Int64())).Take(&notificationService)
	if result.RowsAffected < 1 {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("no notification service entries")
	} else if result.Error != nil {
		return &NSInfoErr, entity.StatusClassInternalServerError, result.Error
	}

	if len(notificationService.NSIPv4) != net.IPv4len {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("invalid slice length for ipv4")
	}

	if len(notificationService.NSIPv6) != net.IPv6len {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("invalid slice length for ipv6")
	}

	ipv4, ok := netip.AddrFromSlice(notificationService.NSIPv4)
	if !ok {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("failed to generate ipv4 address")
	}

	ipv6, ok := netip.AddrFromSlice(notificationService.NSIPv6)
	if !ok {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("failed to generate ipv6 address")
	}

	if !(0 <= notificationService.NSPort && notificationService.NSPort <= 65535) {
		return &NSInfoErr, entity.StatusClassConnectionResolutionFailed, errors.New("invalid port number")
	}

	return &entity.NotificationServiceInfomation{
			ID:   notificationService.NSid,
			IPv4: ipv4,
			IPv6: ipv6,
			Port: uint16(notificationService.NSPort),
		},
		entity.StatusClassSuccess,
		nil
}

func (s *sqlHandler) SetNodeInformation(req entity.ConnectionRequest, commonKey string, expire entity.ExpireDate, cipheType entity.TypeCipherClass) error {
	nodeInfo := entity.NodeInformation{
		FQDN:                "",
		NodeID:              []byte{},
		VirtualIPv4:         []byte{},
		VirtualIPv6:         []byte{},
		ApplicationID:       "",
		ApplicationPort:     0,
		NotificationType:    0,
		CommonKey:           []byte{},
		CommonKeyCipherType: 0,
		CommonKeyLength:     0,
		CommonKeyExpire:     time.Time{},
		CreatedAt:           time.Time{},
		UpdatedAt:           time.Time{},
	}
	result := s.Conn.Limit(1).Where("node_id = ?", fmt.Sprintf("\\x%x", req.BaseHeader.ID)).Find(&nodeInfo)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// insert node information
		// FIXME: 暗号化タイプが固定値になっている
		nodeInfo = entity.NodeInformation{
			NodeID:              req.BaseHeader.ID,
			ApplicationPort:     0,
			NotificationType:    0,
			CommonKey:           []byte(commonKey),
			CommonKeyCipherType: int(cipheType),
			CommonKeyLength:     int(len([]byte(commonKey))),
			CommonKeyExpire:     time.Date(int(expire.Year), time.Month(expire.Month), int(expire.Day), 0, 0, 0, 0, time.Local).AddDate(0, 1, 0),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		config.LogDebug("Set node information", "transaction_id", req.BaseHeader.TransactionID, "information", nodeInfo)

		if err := s.Conn.Create(&nodeInfo).Error; err != nil {
			return err
		}
	} else {
		// update node information
		// FIXME: 暗号化タイプが固定値になっている
		updateNodeInformation := entity.NodeInformation{
			NodeID:              req.BaseHeader.ID,
			CommonKey:           []byte(commonKey),
			CommonKeyCipherType: int(cipheType),
			CommonKeyLength:     int(len([]byte(commonKey))),
			CommonKeyExpire:     time.Date(int(expire.Year), time.Month(expire.Month), int(expire.Day), 0, 0, 0, 0, time.Local).AddDate(0, 1, 0),
			CreatedAt:           nodeInfo.CreatedAt,
			UpdatedAt:           time.Now(),
		}

		config.LogDebug("Updated node information", "transaction_id", req.BaseHeader.TransactionID, "information", updateNodeInformation)

		if err := s.Conn.Model(&nodeInfo).Where("node_id = ?", fmt.Sprintf("\\x%x", req.BaseHeader.ID)).Updates(updateNodeInformation).Error; err != nil {
			return err
		}
	}

	return nil
}
