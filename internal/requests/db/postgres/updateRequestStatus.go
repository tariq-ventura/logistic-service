package requests_db_postgres

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (pc *PostgresClient) UpdateRequestStatus(requestID uuid.UUID, newStatus requests_domain.RequestStatus, reason string) (*requests_domain.Request, *requests_domain.RequestStatusHistory, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(
		pc.ctx,
		"requests.database.postgres",
		map[string]any{
			"db.name":      "requests",
			"db.operation": "update_status",
			"db.type":      "postgresql",
			"request.id":   requestID,
		},
	)

	defer operationSpan.End()

	var updatedRequest requests_domain.Request
	var createdHistory requests_domain.RequestStatusHistory

	transactionError := pc.client.
		WithContext(spanCtx).
		Transaction(func(tx *gorm.DB) error {
			var currentRequest requests_domain.Request

			/*
				SELECT ... FOR UPDATE

				Bloquea la petición mientras se ejecuta la
				transacción, evitando cambios simultáneos.
			*/
			result := tx.
				Clauses(
					clause.Locking{
						Strength: "UPDATE",
					},
				).
				First(
					&currentRequest,
					"id = ?",
					requestID,
				)

			if result.Error != nil {
				return result.Error
			}

			if currentRequest.Status == newStatus {
				return fmt.Errorf(
					"%w: request is already %s",
					validations.ErrInvalidRequestStatusTransition,
					newStatus,
				)
			}

			if !currentRequest.Status.CanTransitionTo(
				newStatus,
			) {
				return fmt.Errorf(
					"%w: %s -> %s",
					validations.ErrInvalidRequestStatusTransition,
					currentRequest.Status,
					newStatus,
				)
			}

			previousStatus := currentRequest.Status

			updateResult := tx.
				Model(&requests_domain.Request{}).
				Where(
					"id = ? AND status = ?",
					requestID,
					previousStatus,
				).
				Updates(
					map[string]any{
						"status":     newStatus,
						"updated_at": time.Now().UTC(),
					},
				)

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected == 0 {
				return validations.ErrConcurrentRequestStatusChange
			}

			createdHistory =
				requests_domain.RequestStatusHistory{
					ID:         uuid.New(),
					RequestID:  requestID,
					FromStatus: previousStatus,
					ToStatus:   newStatus,
					Reason: strings.TrimSpace(
						reason,
					),
					ChangedAt: time.Now().UTC(),
				}

			if err := tx.
				Create(&createdHistory).
				Error; err != nil {

				return err
			}

			if err := tx.
				First(
					&updatedRequest,
					"id = ?",
					requestID,
				).
				Error; err != nil {

				return err
			}

			return nil
		})

	if transactionError != nil {
		switch {
		case errors.Is(
			transactionError,
			gorm.ErrRecordNotFound,
		):
			return nil, nil, &interfaces.Error{
				Error:      "request_not_found",
				Message:    "La petición no existe",
				StatusCode: http.StatusNotFound,
			}

		case errors.Is(
			transactionError,
			validations.ErrInvalidRequestStatusTransition,
		):
			return nil, nil, &interfaces.Error{
				Error: "invalid_status_transition",

				Message: transactionError.Error(),

				StatusCode: http.StatusUnprocessableEntity,
			}

		case errors.Is(
			transactionError,
			validations.ErrConcurrentRequestStatusChange,
		):
			return nil, nil, &interfaces.Error{
				Error: "concurrent_status_change",

				Message: "El estado de la petición cambió " +
					"durante la operación. Consulte nuevamente.",

				StatusCode: http.StatusConflict,
			}

		default:
			pc.logging.LogError(
				"request_status_database_error",
				map[string]any{
					"error": transactionError.Error(),
				},
			)

			return nil, nil, &interfaces.Error{
				Error: "database_error",

				Message: "No se pudo cambiar el estado " +
					"de la petición",

				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	return &updatedRequest, &createdHistory, nil
}
