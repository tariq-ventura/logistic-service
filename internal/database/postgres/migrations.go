package database_postgres

import (
	"context"

	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func (pc *PostgresClient) MigrateDatabase(ctx context.Context) error {
	if err := pc.Client.WithContext(ctx).AutoMigrate(&equipments_domain.Equipment{}, &requests_domain.Request{}, &requests_domain.RequestStatusHistory{}); err != nil {
		return err
	}
	pc.logging.LogInfo("Successfully completed migration in Postgres", nil)
	return nil
}
