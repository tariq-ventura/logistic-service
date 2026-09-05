package requests_db_postgres

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"gorm.io/gorm"
)

func (pc *PostgresClient) UpdateRequest(id uuid.UUID, updates map[string]any) (*requests_domain.Request, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "equipments.database.postgres", map[string]any{
		"db.name":       "requests",
		"db.operation":  "update",
		"db.type":       "postgresql",
		"equipments.Id": id,
	})
	defer operationSpan.End()

	result, err := pc.ListRequestById(id)

	if err != nil {
		return nil, err
	}

	if result.Status != requests_domain.RequestPending {
		return nil, &interfaces.Error{
			Error:      "request_cannot_be_updated",
			Message:    "Solo se pueden modificar peticiones con estado PENDING",
			StatusCode: http.StatusConflict,
		}
	}

	update := pc.client.WithContext(spanCtx).Model(&result).Where("id = ? AND status = ?", id, requests_domain.RequestPending).Updates(updates)

	if update.Error != nil {
		switch {
		case errors.Is(update.Error, gorm.ErrDuplicatedKey):
			return nil, &interfaces.Error{
				Error:      "requests_already_exists",
				Message:    "Ya existe esta peticion",
				StatusCode: http.StatusConflict,
			}

		default:
			pc.logging.LogError("requests_database_error", map[string]any{"error": update.Error.Error()})
			return nil, &interfaces.Error{
				Error:      "database_error",
				Message:    "No se pudo actualizar la peticion",
				StatusCode: http.StatusInternalServerError,
			}
		}
	}
	if update.RowsAffected == 0 {
		return nil, &interfaces.Error{
			Error:      "requests_not_found",
			Message:    "La peticion no existe",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return result, nil
}
