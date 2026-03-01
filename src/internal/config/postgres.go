package config

import (
	"fmt"
	"jk-api/pkg/gorm/audit"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

type logWriter struct{}

func (w logWriter) Printf(format string, args ...interface{}) {
	Logger.WithField("source", "gorm").Infof(format, args...)
}

func PostgresApp() error {
	newLogger := gormLogger.New(
		logWriter{},
		gormLogger.Config{
			SlowThreshold:             0,
			LogLevel:                  gormLogger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(getDsn()), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	DB = db
	if err := db.Use(&audit.AuditLoggerPlugin{}); err != nil {
		return fmt.Errorf("failed to register Audit Logger Plugin: %w", err)
	}

	return nil
}
