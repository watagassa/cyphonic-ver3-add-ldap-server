// Package main contains the main work of the authentication service.
package main

import (
	"crypto/sha512"
	"encoding/hex"
	"os"

	"github.com/Pluslab/cyphonic/ops/domains"
	"github.com/Pluslab/cyphonic/ops/internal"
	"github.com/Pluslab/cyphonic/ops/logger"
)

const isEmailCertficate = 3

func main() {
	cfg, err := internal.Get()

	db, err := internal.GetGormConn(*cfg)
	if err != nil {
		logger.LogErr("database connection error", "error", err)
	}

	a := domains.Accounts{}

	if result := db.Limit(1).Where("email = ?", os.Args[1]).Find(&a); result.Error != nil {
		if result.RowsAffected == 0 {
			// Email: os.Args[1], Password: os.Args[2]
			if len(os.Args) == isEmailCertficate {
				hashedPassword := hashPassword(os.Args[2])

				account := domains.Accounts{
					Email:    os.Args[1],
					Password: hashedPassword,
					State:    1,
				}

				db.Create(&account)
			}
		}
	}
}

// hashPassword hashes plain password to SHA-512.
func hashPassword(plainPassword string) (hashedPassword string) {
	b := []byte(plainPassword)
	sha512 := sha512.Sum512(b)

	return hex.EncodeToString(sha512[:])
}
