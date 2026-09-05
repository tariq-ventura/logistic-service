package requests_db_postgres

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func (pc *PostgresClient) ListRequestStatusHistory(id uuid.UUID) ([]requests_domain.RequestStatusHistory, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "requests.database.postgres", map[string]any{
		"db.name":      "requests",
		"db.operation": "list",
		"db.type":      "postgresql",
		"requests.Id":  id,
	})
	defer operationSpan.End()

	_, err := pc.ListRequestById(id)

	if err != nil {
		return nil, err
	}

	var history []requests_domain.RequestStatusHistory

	result := pc.client.WithContext(spanCtx).Where("request_id = ?", id).Order("changed_at DESC").Find(&history)

	if result.Error != nil {
		pc.logging.LogError("database_error", map[string]any{"error": result.Error.Error()})
		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo consultar el historial",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return history, nil
}
