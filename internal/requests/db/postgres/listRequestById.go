package requests_db_postgres

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"gorm.io/gorm"
)

func (pc *PostgresClient) ListRequestById(id uuid.UUID) (*requests_domain.Request, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "requests.database.postgres", map[string]any{
		"db.name":      "requests",
		"db.operation": "list",
		"db.type":      "postgresql",
		"requests.Id":  id,
	})
	defer operationSpan.End()

	var request requests_domain.Request

	result := pc.client.WithContext(spanCtx).First(&request, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, &interfaces.Error{
			Error:      "request_not_found",
			Message:    "La peticion no existe",
			StatusCode: http.StatusNotFound,
		}
	}

	if result.Error != nil {
		pc.logging.LogError("database_error", map[string]any{"error": result.Error.Error()})
		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo consultar la peticion",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &request, nil
}
