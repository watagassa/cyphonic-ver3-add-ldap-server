package infrastructure

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/Pluslab/cyphonic/asd/entity"
	"github.com/Pluslab/cyphonic/asd/infrastructure/config"
	"github.com/Pluslab/cyphonic/asd/usecase/repository"
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

func (s *sqlHandler) GetDevice(ctx context.Context, lreq entity.LoginRequest, serverCert, rootCert []byte) (entity.Device, entity.StatusClass, error) {
	devicePassword := entity.DevicePassword{
		DeviceID: "",
		Password: "",
	}

	device := entity.Device{
		Version:                    "",
		DeviceID:                   "",
		AccountID:                  0,
		DeviceName:                 "",
		FQDN:                       "",
		NodeManagementAreaID:       0,
		DeviceTypeID:               0,
		GroupProfile:               0,
		AdapterFlag:                false,
		GeneralNodeFlag:            false,
		InterNodeCertificationFlag: false,
		Status:                     0,
	}

	switch lreq.AuthType {
	case entity.TypeDeviceIDPassword:
		var count int64

		err := s.Conn.Limit(1).Where("device_id = ? AND password = ?", string(lreq.DeviceID), fmt.Sprintf("%x", lreq.Password)).Find(&devicePassword).Count(&count).Error
		if err != nil {
			return device, entity.StatusClassInternalServerError, fmt.Errorf("can't get device's account: %w", err)
		} else if count < 1 {
			return device, entity.StatusClassAuthenticationFailed, fmt.Errorf("authentication failed: %w", err)
		}

		s.Conn.Limit(1).Where("device_id = ?", devicePassword.DeviceID).Find(&device)

	case entity.TypeDigitalAuthentication:
		cert, err := lreq.GetDigitalCertificate()
		if err != nil {
			return device, entity.StatusClassInternalServerError, fmt.Errorf("can't get digital certificate: %w", err)
		}

		if err = verifyDigitalCertificate(cert, serverCert, rootCert); err != nil {
			return device, entity.StatusClassAuthenticationFailed, fmt.Errorf("failed login authentication of digital certificate: %w", err)
		}

		if result := s.Conn.Limit(1).Where("fqdn = ?", cert.Subject.CommonName).Find(&device); result.Error != nil {
			return device, entity.StatusClassInternalServerError, fmt.Errorf("can't get device's information: %w", result.Error)
		}

	case entity.TypeSingleSignOn:
	}

	return device, entity.StatusClassSuccess, nil
}

// verifyDigitalCertificate verify of digital certificate.
func verifyDigitalCertificate(cert *x509.Certificate, serverCert, rootCert []byte) error {
	rootsPool := x509.NewCertPool()
	intermediasPool := x509.NewCertPool()

	if ok := intermediasPool.AppendCertsFromPEM(serverCert); !ok {
		return fmt.Errorf("can't append server certificate")
	}

	if ok := rootsPool.AppendCertsFromPEM(rootCert); !ok {
		return fmt.Errorf("the certificate is incorrect")
	}

	opts := x509.VerifyOptions{
		DNSName:                   "",
		Intermediates:             intermediasPool,
		Roots:                     rootsPool,
		CurrentTime:               time.Time{},
		KeyUsages:                 []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		MaxConstraintComparisions: 0,
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("can't verify: %w", err)
	}

	return nil
}

func (s *sqlHandler) GetVirtualIPAddress(device entity.Device) (entity.VirtualIPAddress, error) {
	virtualIPAddress := entity.VirtualIPAddress{
		DeviceID:           "",
		VirtualIPv4:        []byte{},
		VirtualIPv4Netmask: []byte{},
		VirtualIPv6:        []byte{},
		VirtualIPv6Prefix:  []byte{},
	}

	if result := s.Conn.Limit(1).Where("device_id = ?", device.DeviceID).Find(&virtualIPAddress); result.Error != nil {
		return virtualIPAddress, result.Error
	}

	return virtualIPAddress, nil
}

func (s *sqlHandler) SetNodeInformation(ctx context.Context, device entity.Device, virtualIP entity.VirtualIPAddress, req *entity.LoginRequest, nodeID []byte) error {
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
		DeviceTypeID:       0,
		GroupProfile:       0,
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
	}
	result := s.Conn.Limit(1).Where("node_id = ?", fmt.Sprintf("\\x%x", nodeID)).Find(&nodeInfo)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// insert node information
		nodeInfo = entity.NodeInformation{
			Version:            string(req.BaseHeader.Version),
			FQDN:               string(device.FQDN),
			DeviceID:           string(device.DeviceID),
			NodeID:             nodeID,
			VirtualIPv4:        virtualIP.VirtualIPv4,
			VirtualIPv4Netmask: virtualIP.VirtualIPv4Netmask,
			VirtualIPv6:        virtualIP.VirtualIPv6,
			VirtualIPv6Prefix:  virtualIP.VirtualIPv6Prefix,
			ApplicationID:      fmt.Sprintf("%x", req.BaseHeader.ID),
			ApplicationPort:    0,
			NotificationType:   0,
			DeviceTypeID:       device.DeviceTypeID,
			GroupProfile:       device.GroupProfile,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		config.LogDebug("Set node information", "transaction_id", req.BaseHeader.TransactionID, "information", nodeInfo)

		if err := s.Conn.Create(&nodeInfo).Error; err != nil {
			return err
		}
	} else {
		// update node information
		updateNodeInformation := entity.NodeInformation{
			Version:            string(req.BaseHeader.Version),
			FQDN:               string(device.FQDN),
			DeviceID:           string(device.DeviceID),
			NodeID:             nodeID,
			VirtualIPv4:        virtualIP.VirtualIPv4,
			VirtualIPv4Netmask: virtualIP.VirtualIPv4Netmask,
			VirtualIPv6:        virtualIP.VirtualIPv6,
			VirtualIPv6Prefix:  virtualIP.VirtualIPv6Prefix,
			ApplicationID:      fmt.Sprintf("%x", req.BaseHeader.ID),
			ApplicationPort:    0,
			NotificationType:   0,
			DeviceTypeID:       device.DeviceTypeID,
			GroupProfile:       device.GroupProfile,
			CreatedAt:          nodeInfo.CreatedAt,
			UpdatedAt:          time.Now(),
		}

		config.LogDebug("Updated node information", "transaction_id", req.BaseHeader.TransactionID, "information", updateNodeInformation)

		if err := s.Conn.Model(&nodeInfo).Where("node_id = ?", fmt.Sprintf("\\x%x", nodeID)).Updates(updateNodeInformation).Error; err != nil {
			return err
		}
	}

	return nil
}
