package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config is a struct for environment variables.
type Config struct {
	ServiceType      string
	DatabaseUser     string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseSSLMode  string
	DatabaseTimeZone string
	FQDN             string
	NSIPv4           string
	NSIPv6           string
	Port             string
	DSIPv4           string
	DSIPv6           string
	DSFQDN           string
	DSPort           string
	DebugLogFilePath string
	ErrorLogFilePath string
	DebugMode        bool
	DBDebugMode      bool
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
		{&cfg.NSIPv4, "NS_IPV4"},
		{&cfg.NSIPv6, "NS_IPV6"},
		{&cfg.Port, "PORT"},
		{&cfg.DSIPv4, "DS_IPV4"},
		{&cfg.DSIPv6, "DS_IPV6"},
		{&cfg.DSFQDN, "DS_FQDN"},
		{&cfg.DSPort, "DS_PORT"},
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
