package database_postgres

import (
	"context"

	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func (pc *PostgresClient) MigrateDatabase(ctx context.Context) error {
	if err := pc.Client.WithContext(ctx).
		Exec(`CREATE EXTENSION IF NOT EXISTS vector`).
		Error; err != nil {
		return err
	}

	if err := pc.Client.WithContext(ctx).AutoMigrate(
		&requests_domain.Request{},
		&requests_domain.RequestStatusHistory{},
		&assignments_domain.Assignment{},
	); err != nil {
		return err
	}

	if err := pc.Client.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS logistics_request_embeddings (
			request_id UUID PRIMARY KEY
				REFERENCES logistics_requests(id)
				ON DELETE CASCADE,

			provider VARCHAR(30) NOT NULL,
			model VARCHAR(100) NOT NULL,

			content TEXT NOT NULL,
			embedding VECTOR(768) NOT NULL,

			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_request_embeddings_provider_model
		ON logistics_request_embeddings(provider, model);

		CREATE INDEX IF NOT EXISTS idx_logistics_request_embeddings_hnsw
		ON logistics_request_embeddings
		USING hnsw (embedding vector_cosine_ops);
	`).Error; err != nil {
		return err
	}

	if err := pc.Client.WithContext(ctx).Exec(`
		ALTER TABLE logistics_request_embeddings
			ADD COLUMN IF NOT EXISTS provider VARCHAR(30),
			ADD COLUMN IF NOT EXISTS model VARCHAR(100)
	`).Error; err != nil {
		return err
	}

	pc.logging.LogInfo(
		"Successfully completed migration in Postgres",
		nil,
	)

	return nil
}
