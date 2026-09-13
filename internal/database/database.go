package database

import (
	"context"
	"errors"

	database_postgres "github.com/tariq-ventura/logistic-service/internal/database/postgres"
	equipments_db "github.com/tariq-ventura/logistic-service/internal/equipments/db"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requests_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

type Database struct {
	Equipments equipments_db.IEquipmentsDB
	Requests   requests_db.IRequestsDB
}
type IDatabase interface {
	MigrateDatabase(ctx context.Context) error
}

var SetupDatabase = func(ctx context.Context, l logging.ILogging, t interfaces.ITrace) (*Database, IDatabase, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, nil, err
	}
	if dbType != "postgresql" {
		return nil, nil, errors.New("unsupported database backend")
	}
	db, err := database_postgres.SetupPostgres(l)
	if err != nil {
		return nil, nil, err
	}
	equipments, err := equipments_db.NewDatabase(ctx, l, t, db.Client)
	if err != nil {
		return nil, nil, err
	}
	requests, err := requests_db.NewDatabase(ctx, l, t, db.Client)
	if err != nil {
		return nil, nil, err
	}
	return &Database{Equipments: equipments, Requests: requests}, db, nil
}
