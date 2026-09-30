// Package main contains the main work of the authentication service.
package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"os"
	"time"

	"github.com/Pluslab/cyphonic/ops/certificate"
	"github.com/Pluslab/cyphonic/ops/domains"
	"github.com/Pluslab/cyphonic/ops/internal"
	"github.com/Pluslab/cyphonic/ops/logger"
	"github.com/google/uuid"
)

const (
	isPasswordCertficate = 5
	isDigitalCertficate  = 4
	certificateExpire    = 10
)

func main() {
	var byteFQDN, byteDeviceID bytes.Buffer

	cfg, err := internal.Get()
	db, err := internal.GetGormConn(*cfg)
	if err != nil {
		logger.LogErr("database connection error", "error", err)
	}

	a := domains.Accounts{}
	nmArea := domains.NodeManagementArea{}
	d := domains.Devices{}
	dc := domains.DigitalCertificates{}

	hashedPassword := hashPassword(os.Args[2])

	if result := db.Limit(1).Where("email = ? AND password = ?", os.Args[1], hashedPassword).Find(&a); result.Error == nil {
		if result := db.Limit(1).Where("area = ?", "asia").Find(&nmArea); result.Error == nil {
			if result := db.Limit(1).Where("device_name = ?", os.Args[3]).Find(&d); result.Error != nil {
				if result.RowsAffected == 0 {

					di, err := generateDeviceID()
					if err != nil {
						logger.LogErr("Failed to generate Device ID", "error", err)
						return
					}

					hashID := hashID(di + os.Args[1])
					byteDeviceID.WriteString(hashID)

					byteFQDN.WriteString(hashID)
					byteFQDN.WriteString(".")
					byteFQDN.WriteString(nmArea.FQDN)

					device := domains.Devices{
						Version:                    "1.0",
						DeviceID:                   byteDeviceID.String(),
						AccountID:                  a.ID,
						DeviceName:                 os.Args[3],
						FQDN:                       byteFQDN.String(),
						NodeManagementAreaID:       nmArea.ID,
						DeviceType:                 1,
						AdapterFlag:                false,
						GeneralNodeFlag:            false,
						InterNodeCertificationFlag: false,
						Status:                     0,
					}

					db.Create(&device)

					if len(os.Args) == isDigitalCertficate {

						result := db.Limit(1).Order("id DESC").Find(&dc)
						id := 1
						if result.Error == nil {
							id = dc.ID + 1
						}
						generateClientCertificate(id, os.Args[3], os.Args[1], device.FQDN)

						certificate := domains.DigitalCertificates{
							DeviceID: byteDeviceID.String(),
							Status:   0,
							Expire:   time.Now().AddDate(certificateExpire, 0, 0),
						}

						db.Create(&certificate)
					} else if len(os.Args) == isPasswordCertficate {

						hashPassword := hashPassword(os.Args[4])

						password := domains.DevicePasswords{
							DeviceID: byteDeviceID.String(),
							Password: hashPassword,
						}

						db.Create(&password)
					}
				}
			}
		}
	} else {
		logger.LogErr("Failed to create cyphonic device", "error")
		return
	}
}

// hashID hashes plain id to SHA-224.
func hashID(plainEmail string) (hashedEmail string) {
	b := []byte(plainEmail)
	sha224 := sha256.Sum224(b)

	return hex.EncodeToString(sha224[:])
}

// hashPassword hashes plain password to SHA-512.
func hashPassword(plainPassword string) (hashedPassword string) {
	b := []byte(plainPassword)
	sha512 := sha512.Sum512(b)

	return hex.EncodeToString(sha512[:])
}

// generateDeviceID is to generate Device ID.
func generateDeviceID() (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		logger.LogErr("Failed to generate UUID for ID", "error", err)
		return "", err
	}

	id, err := uuid.MarshalBinary()
	if err != nil {
		logger.LogErr("Failed to encode to binary format for ID", "error", err)
		return "", err
	}

	return string(id), nil
}

// generateClientCertificate generates a client certificate.
func generateClientCertificate(serialNum int, commonName, nodeEmail, fqdn string) {
	if _, err := certificate.GenerateCYPClientCert(serialNum, commonName, nodeEmail, fqdn); err != nil {
		logger.LogErr("Failed to get the certificate information", "error", err)
	}
}
