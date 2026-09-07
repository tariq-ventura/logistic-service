package assignments_db_postgres

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (pc *PostgresClient) CreateAssignment(assignment assignments_domain.Assignment, reason string) *interfaces.Error {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "equipments.database.postgres", map[string]any{
		"db.name":               "assignments",
		"db.operation":          "insert",
		"db.type":               "postgresql",
		"assignments.createdAt": assignment.CreatedAt,
	})
	defer operationSpan.End()

	now := time.Now().UTC()

	transactionError := pc.client.
		WithContext(spanCtx).
		Transaction(func(tx *gorm.DB) error {
			var request requests_domain.Request

			result := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&request, "id = ?", assignment.RequestID)

			if result.Error != nil {
				return result.Error
			}

			if request.Status != requests_domain.RequestPending {
				return fmt.Errorf(
					"%w: current status is %s",
					validations.ErrRequestNotPending,
					request.Status,
				)
			}

			if err := tx.Create(&assignment).Error; err != nil {
				return err
			}

			updateResult := tx.
				Model(&requests_domain.Request{}).
				Where(
					"id = ? AND status = ?",
					assignment.RequestID,
					requests_domain.RequestPending,
				).
				Updates(map[string]any{
					"status":     requests_domain.RequestAssigned,
					"updated_at": now,
				})

			if updateResult.Error != nil {
				return updateResult.Error
			}

			if updateResult.RowsAffected == 0 {
				return validations.ErrConcurrentAssignment
			}

			history := requests_domain.RequestStatusHistory{
				ID:         uuid.New(),
				RequestID:  assignment.RequestID,
				FromStatus: requests_domain.RequestPending,
				ToStatus:   requests_domain.RequestAssigned,
				Reason:     strings.TrimSpace(reason),
				ChangedAt:  now,
			}

			return tx.Create(&history).Error
		})

	if transactionError != nil {
		switch {
		case errors.Is(transactionError, gorm.ErrRecordNotFound):
			return &interfaces.Error{
				Error:      "request_not_found",
				Message:    "La petición no existe",
				StatusCode: http.StatusNotFound,
			}

		case errors.Is(transactionError, gorm.ErrDuplicatedKey):
			return &interfaces.Error{
				Error:      "request_already_assigned",
				Message:    "La petición ya tiene una asignación",
				StatusCode: http.StatusConflict,
			}

		case errors.Is(transactionError, validations.ErrRequestNotPending):
			return &interfaces.Error{
				Error:      "request_not_pending",
				Message:    "Solo se pueden asignar peticiones PENDING",
				StatusCode: http.StatusConflict,
			}

		case errors.Is(transactionError, validations.ErrConcurrentAssignment):
			return &interfaces.Error{
				Error:      "concurrent_assignment",
				Message:    "La petición cambió durante la asignación",
				StatusCode: http.StatusConflict,
			}

		default:
			pc.logging.LogError(
				"assignment_database_error",
				map[string]any{"error": transactionError.Error()},
			)

			return validations.AssignmentDatabaseError()
		}
	}

	return nil
}
