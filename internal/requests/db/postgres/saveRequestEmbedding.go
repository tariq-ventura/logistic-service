package requests_db_postgres

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
)

func (pc *PostgresClient) SaveRequestEmbedding(
	ctx context.Context,
	requestID uuid.UUID,
	provider string,
	model string,
	content string,
	embedding []float32,
) *interfaces.Error {
	if len(embedding) != 768 {
		return &interfaces.Error{
			Error:      "invalid_embedding",
			Message:    "El embedding debe contener 768 dimensiones",
			StatusCode: http.StatusInternalServerError,
		}
	}

	result := pc.client.WithContext(ctx).Exec(`
		INSERT INTO logistics_request_embeddings (
			request_id,
			provider,
			model,
			content,
			embedding,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, NOW())
		ON CONFLICT (request_id)
		DO UPDATE SET
			provider = EXCLUDED.provider,
			model = EXCLUDED.model,
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			updated_at = NOW()
	`,
		requestID,
		provider,
		model,
		content,
		pgvector.NewVector(embedding),
	)

	if result.Error != nil {
		return &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo guardar el índice semántico",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return nil
}
