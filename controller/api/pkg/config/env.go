package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseUser     string
	DatabaseProtocol string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseSSLMode  string
	DatabaseTimeZone string
	FQDN             string
	Port             string
	DebugMode        bool
	DBDebugMode      bool
}

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
		{&cfg.DatabaseUser, "DATABASE_USER"},
		{&cfg.DatabaseHost, "DATABASE_HOST"},
		{&cfg.DatabasePort, "DATABASE_PORT"},
		{&cfg.DatabaseName, "DATABASE_NAME"},
		{&cfg.DatabaseSSLMode, "DATABASE_SSL_MODE"},
		{&cfg.DatabaseTimeZone, "DATABASE_TIME_ZONE"},
		{&cfg.FQDN, "FQDN"},
		{&cfg.Port, "PORT"},
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
