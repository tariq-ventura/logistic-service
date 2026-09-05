package database_postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/tariq-ventura/logistic-service/internal/logging"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresClient struct {
	Client  *gorm.DB
	logging logging.ILogging
}

var SetupPostgres = func(logs logging.ILogging) (*PostgresClient, error) {
	dsn, err := validations.RequiredEnv("DB_STRING")
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting connection pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf(
			"checking PostgreSQL connection: %w",
			err,
		)
	}

	logs.LogInfo("PostgreSQL connected successfully", nil)

	return &PostgresClient{
		Client:  db,
		logging: logs,
	}, nil
}
