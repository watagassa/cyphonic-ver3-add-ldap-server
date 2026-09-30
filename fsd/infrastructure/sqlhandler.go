package infrastructure

import (
	"context"
	"fmt"

	"github.com/Pluslab/cyphonic/fsd/entity"
	"github.com/Pluslab/cyphonic/fsd/usecase/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type sqlHandler struct {
	Conn *gorm.DB
	FQDN string
}

func NewSQLHandler(host, user, name, port, sslMode, timeZone, fqdn string) (repository.SQLHandler, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, name, port, sslMode, timeZone)

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
		return nil, fmt.Errorf("can't open gorm: %w", err)
	}

	return &sqlHandler{
		Conn: conn,
		FQDN: fqdn,
	}, nil
}

func (s *sqlHandler) DeleteNodeInformation(ctx context.Context, req entity.FinalizationRequest) error {
	nodeInfo := entity.NodeInformation{}

	result := s.Conn.Limit(1).Where("node_id = ?", fmt.Sprintf("\\x%x", req.BaseHeader.ID)).First(&nodeInfo)
	if result.Error != nil {
		return result.Error
	}

	// Remove child element
	if err := s.Conn.Where("node_id = ?", fmt.Sprintf("\\x%x", req.BaseHeader.ID)).Delete(&entity.NodeAddress{}).Error; err != nil {
		return err
	}

	// Remove parent element
	// TODO: 親が削除されたら子も削除されるようなテーブル設定をする
	deleteResult := s.Conn.Where("node_id = ?", fmt.Sprintf("\\x%x", req.BaseHeader.ID)).Delete(&entity.NodeInformation{})
	if deleteResult.Error != nil {
		return deleteResult.Error
	}

	return nil
}
