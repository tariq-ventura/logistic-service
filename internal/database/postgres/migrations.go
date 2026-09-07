package database_postgres

import (
	"context"

	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func (pc *PostgresClient) MigrateDatabase(ctx context.Context) error {
	err := pc.Client.AutoMigrate(
		&requests_domain.Request{},
		&requests_domain.RequestStatusHistory{},
		&assignments_domain.Assignment{},
	)

	if err != nil {
		return err
	}

	pc.logging.LogInfo("Successfully completed migration in Postgres", nil)
	return nil
}
