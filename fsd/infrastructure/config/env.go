package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config is a struct for environment variables.
type Config struct {
	ServiceType         string
	DatabaseUser        string
	DatabaseHost        string
	DatabasePort        string
	DatabaseName        string
	DatabaseSSLMode     string
	DatabaseTimeZone    string
	FQDN                string
	TLSPort           	string
	QUICPort			string
	RootCertName        string
	AccessKey           string
	SecretKey           string
	Region              string
	EndPoint            string
	BucketName          string
	CertificateLocation string
	DebugLogFilePath    string
	ErrorLogFilePath    string
	DebugMode           bool
	DBDebugMode         bool
}

// Get gets a value from environment variables.
func Get() (*Config, error) {
	var (
		cfg         Config
		missed      []string
		debugMode   string
		dbDebugMode string
	)

	for _, prop := range []struct {
		field *string
		name  string
	}{
		{&cfg.ServiceType, "SERVICE_TYPE"},
		{&cfg.DatabaseUser, "DATABASE_USER"},
		{&cfg.DatabaseHost, "DATABASE_HOST"},
		{&cfg.DatabasePort, "DATABASE_PORT"},
		{&cfg.DatabaseName, "DATABASE_NAME"},
		{&cfg.DatabaseSSLMode, "DATABASE_SSL_MODE"},
		{&cfg.DatabaseTimeZone, "DATABASE_TIME_ZONE"},
		{&cfg.FQDN, "FQDN"},
		{&cfg.TLSPort, "TLS_PORT"},
		{&cfg.QUICPort, "QUIC_PORT"},
		{&cfg.RootCertName, "ROOT_CERTIFICATE_NAME"},
		{&cfg.AccessKey, "ACCESS_KEY"},
		{&cfg.SecretKey, "SECRET_KEY"},
		{&cfg.Region, "REGION"},
		{&cfg.EndPoint, "END_POINT"},
		{&cfg.BucketName, "BUCKET_NAME"},
		{&cfg.CertificateLocation, "CERTIFICATE_LOCATION"},
		{&cfg.DebugLogFilePath, "DEBUG_LOG_FILE_PATH"},
		{&cfg.ErrorLogFilePath, "ERROR_LOG_FILE_PATH"},
		{&debugMode, "DEBUG_MODE"},
		{&dbDebugMode, "DB_DEBUG_MODE"},
	} {
		v := os.Getenv(prop.name)
		*prop.field = v

		if v == "" {
			missed = append(missed, prop.name)
		}
	}

	if len(missed) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missed)
	}

	isDebugMode, err := strconv.ParseBool(debugMode)
	if err != nil {
		return nil, fmt.Errorf("parse error debugMode: %w", err)
	}

	cfg.DebugMode = isDebugMode

	isDBDebugMode, err := strconv.ParseBool(dbDebugMode)
	if err != nil {
		return nil, fmt.Errorf("parse error db debugMode: %w", err)
	}

	cfg.DBDebugMode = isDBDebugMode

	return &cfg, nil
}
