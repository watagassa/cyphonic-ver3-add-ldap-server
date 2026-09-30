package internal

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/xerrors"
)

type Config struct {
	DatabaseUser     string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseSSLMode  string
	DatabaseTimeZone string
	FQDN             string
	Port             string
}

func Get() (*Config, error) {
	envLoad()

	var (
		cfg    Config
		missed []string
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
	} {
		v := os.Getenv(prop.name)
		*prop.field = v

		if v == "" {
			missed = append(missed, prop.name)
		}
	}

	if len(missed) > 0 {
		return nil, xerrors.Errorf("missing required environment variables: %w", missed)
	}

	return &cfg, nil
}

// envLoad loads .env file.
func envLoad() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}
