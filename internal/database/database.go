package database

import (
	"context"
	"errors"

	assignmets_db "github.com/tariq-ventura/logistic-service/internal/assigments/db"
	database_postgres "github.com/tariq-ventura/logistic-service/internal/database/postgres"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requets_db "github.com/tariq-ventura/logistic-service/internal/requests/db"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

type Database struct {
	Requests    requets_db.IRequestsDB
	Assignments assignmets_db.IAssignmentsDB
}

type IDatabase interface {
	MigrateDatabase(ctx context.Context) error
}

var SetupDatabase = func(ctx context.Context, l logging.ILogging, t interfaces.ITrace) (*Database, IDatabase, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, nil, err
	}

	switch dbType {
	case "postgresql":
		l.LogInfo("Database selected", map[string]any{"db_type": dbType})
		db, err := database_postgres.SetupPostgres(l)

		if err != nil {
			return nil, nil, err
		}

		request, err := requets_db.NewDatabase(ctx, l, t, db.Client)

		if err != nil {
			return nil, nil, err
		}

		assigment, err := assignmets_db.NewDatabase(ctx, l, t, db.Client)

		if err != nil {
			return nil, nil, err
		}

		return &Database{
			Requests:    request,
			Assignments: assigment,
		}, db, err
	default:
		return nil, nil, errors.New("unsupported database backend")
	}
}
