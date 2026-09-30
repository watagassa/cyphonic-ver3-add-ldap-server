package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EnvError struct {
	Missed []string
}

func (e *EnvError) Error() string {
	return fmt.Sprintf("missing required environment variables: %v", e.Missed)
}

// Config is a struct for environment variables.
type Config struct {
	ServiceType      string
	FQDN             string
	Port             string
	DatabaseUser     string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseSSLMode  string
	DatabaseTimeZone string
	RedisAddresses   []string
	DebugLogFilePath string
	ErrorLogFilePath string
	CacheClusterMode bool
	DebugMode        bool
	DBDebugMode      bool
}

// Get gets a value from environment variables.
func Get() (*Config, error) {
	var (
		cfg              Config
		missed           []string
		cacheClusterMode string
		debugMode        string
		dbDebugMode      string
		redisAddresses   string
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
		{&redisAddresses, "REDIS_ADDRESSES"},
		{&cfg.FQDN, "FQDN"},
		{&cfg.Port, "PORT"},
		{&cfg.DebugLogFilePath, "DEBUG_LOG_FILE_PATH"},
		{&cfg.ErrorLogFilePath, "ERROR_LOG_FILE_PATH"},
		{&cacheClusterMode, "CACHE_CLUSTER_MODE"},
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
		return nil, &EnvError{Missed: missed}
	}

	cfg.RedisAddresses = strings.Split(redisAddresses, ",")

	isCacheClusterMode, err := strconv.ParseBool(cacheClusterMode)
	if err != nil {
		return nil, fmt.Errorf("parse error cacheClusterMode: %w", err)
	}

	cfg.CacheClusterMode = isCacheClusterMode

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
