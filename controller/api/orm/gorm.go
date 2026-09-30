// Package orm provides Object-Relational Mapping for database handlers.
package orm

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/controller/api/adapter/gateway"
	"github.com/Pluslab/cyphonic/controller/api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CustomGormClientFactory struct {
	Cfg *config.Config
}

func (g *CustomGormClientFactory) NewClient(ctx context.Context) (*gateway.Client, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		g.Cfg.DatabaseHost, g.Cfg.DatabaseUser, g.Cfg.DatabaseName,
		g.Cfg.DatabasePort, g.Cfg.DatabaseSSLMode, g.Cfg.DatabaseTimeZone)

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
		return nil, fmt.Errorf("failed to open gorm: %w", err)
	}

	return &gateway.Client{
		Conn: conn,
	}, nil
}
