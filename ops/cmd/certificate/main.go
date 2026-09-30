// Package main contains the main work of the authentication service.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Pluslab/cyphonic/ops/certificate"
	"github.com/Pluslab/cyphonic/ops/domains"
	"github.com/Pluslab/cyphonic/ops/internal"
	"github.com/Pluslab/cyphonic/ops/logger"
)

const (
	isCreateCA        = 1
	isCreateNode      = 3
	certificateExpire = 10
)

func main() {
	cfg, err := internal.Get()

	db, err := internal.GetGormConn(*cfg)
	if err != nil {
		logger.LogErr("database connection error", "error", err)
	}
	d := domains.Devices{}
	dc := domains.DigitalCertificates{}

	switch len(os.Args) {
	case isCreateCA:
		_, err := certificate.InitCert()
		if err != nil {
			logger.LogErr("Failed to get the server certificate information", "error", err)
		}
	case isCreateNode:
		if result := db.Debug().Limit(1).Where("device_name=?", os.Args[2]).Find(&d); result.Error == nil {
			if result := db.Debug().Limit(1).Where("device_id=?", d.DeviceID).Find(&dc); result.Error == nil {
				if result.RowsAffected == 0 {
					id := 1
					if result := db.Limit(1).Order("id DESC").Find(&dc); result.Error == nil {
						id = dc.ID + 1
					}

					err := generateCYPCertificate(id, os.Args[2], os.Args[1], d.FQDN)
					if err != nil {
						logger.LogErr("Failed to generate certificate", "error", err)
						return
					}

					cert := domains.DigitalCertificates{
						DeviceID: d.DeviceID,
						Status:   0,
						Expire:   time.Now().AddDate(certificateExpire, 0, 0),
					}

					db.Create(&cert)
				} else {
					fmt.Println(result.Error)
					err := generateCYPCertificate(dc.ID, os.Args[2], os.Args[1], d.FQDN)
					if err != nil {
						logger.LogErr("Failed to generate certificate", "error", err)
						return
					}

					db.Model(dc).Updates(
						domains.DigitalCertificates{
							Status: 0,
							Expire: time.Now().AddDate(certificateExpire, 0, 0),
						})
				}
			}
		}
	}
}

// generateCYPCertificate generates a cyphonic certificate.
func generateCYPCertificate(serialNum int, commonName, nodeEmail, fqdn string) error {
	if _, err := certificate.GenerateCYPClientCert(serialNum, commonName, nodeEmail, fqdn); err != nil {
		logger.LogErr("Failed to get the certificate information", "error", err)
		return err
	}
	return nil
}
