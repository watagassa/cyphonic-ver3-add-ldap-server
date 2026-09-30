// Package certificate contains the certficate of TLS connetction.
package certificate

import (
	"crypto/tls"

	"github.com/Pluslab/cyphonic/ops/internal"
)

// InitCert return certificate information.
func InitCert() (serverTLSConf *tls.Config, err error) {
	cfg, err := internal.Get()

	_, _, err = GenerateRootCA()
	if err != nil {
		return nil, err
	}
	// Generate AS's certificate
	_, err = GenerateServerCert(cfg.FQDN)
	if err != nil {
		return nil, err
	}

	// Generate PS's Server Certificate
	// FIX ME: 引数が固定値になっている
	_, err = GenerateServerCert("ps.local.cyphonic.org")
	if err != nil {
		return nil, err
	}

	// Generate CMS's certificate
	_, err = GenerateServerCert("cms.local.cyphonic.org")
	if err != nil {
		return nil, err
	}

	// Generate FS's certificate
	_, err = GenerateServerCert("fs.local.cyphonic.org")
	if err != nil {
		return nil, err
	}

	return serverTLSConf, nil
}
