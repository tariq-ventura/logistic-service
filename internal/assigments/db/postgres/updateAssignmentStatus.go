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

func (pc *PostgresClient) UpdateAssignmentStatus(
	assignmentID uuid.UUID,
	newStatus assignments_domain.AssignmentStatus,
	reason string,
) (*assignments_domain.Assignment, *interfaces.Error) {
	now := time.Now().UTC()
	var updatedAssignment assignments_domain.Assignment

	transactionError := pc.client.
		WithContext(pc.ctx).
		Transaction(func(tx *gorm.DB) error {
			var assignment assignments_domain.Assignment

			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&assignment, "id = ?", assignmentID).
				Error; err != nil {
				return err
			}

			if assignment.Status != assignments_domain.AssignmentActive {
				return fmt.Errorf(
					"%w: current status is %s",
					validations.ErrAssignmentNotActive,
					assignment.Status,
				)
			}

			var request requests_domain.Request

			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&request, "id = ?", assignment.RequestID).
				Error; err != nil {
				return err
			}

			if request.Status != requests_domain.RequestAssigned {
				return fmt.Errorf(
					"%w: current status is %s",
					validations.ErrRequestNotAssigned,
					request.Status,
				)
			}

			requestStatus := requests_domain.RequestCompleted

			assignmentUpdates := map[string]any{
				"status":        newStatus,
				"status_reason": strings.TrimSpace(reason),
				"updated_at":    now,
			}

			if newStatus == assignments_domain.AssignmentCompleted {
				assignmentUpdates["completed_at"] = now
			} else {
				requestStatus = requests_domain.RequestCancelled
				assignmentUpdates["cancelled_at"] = now
			}

			result := tx.
				Model(&assignments_domain.Assignment{}).
				Where(
					"id = ? AND status = ?",
					assignmentID,
					assignments_domain.AssignmentActive,
				).
				Updates(assignmentUpdates)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return validations.ErrConcurrentAssignment
			}

			requestUpdate := tx.
				Model(&requests_domain.Request{}).
				Where(
					"id = ? AND status = ?",
					request.ID,
					requests_domain.RequestAssigned,
				).
				Updates(map[string]any{
					"status":     requestStatus,
					"updated_at": now,
				})

			if requestUpdate.Error != nil {
				return requestUpdate.Error
			}

			if requestUpdate.RowsAffected == 0 {
				return validations.ErrConcurrentAssignment
			}

			history := requests_domain.RequestStatusHistory{
				ID:         uuid.New(),
				RequestID:  request.ID,
				FromStatus: requests_domain.RequestAssigned,
				ToStatus:   requestStatus,
				Reason:     strings.TrimSpace(reason),
				ChangedAt:  now,
			}

			if err := tx.Create(&history).Error; err != nil {
				return err
			}

			return tx.
				First(&updatedAssignment, "id = ?", assignmentID).
				Error
		})

	if transactionError != nil {
		switch {
		case errors.Is(transactionError, gorm.ErrRecordNotFound):
			return nil, validations.AssignmentNotFound()

		case errors.Is(transactionError, validations.ErrAssignmentNotActive):
			return nil, &interfaces.Error{
				Error:      "assignment_not_active",
				Message:    "La asignación ya fue finalizada",
				StatusCode: http.StatusUnprocessableEntity,
			}

		case errors.Is(transactionError, validations.ErrRequestNotAssigned):
			return nil, &interfaces.Error{
				Error:      "request_not_assigned",
				Message:    "La petición no se encuentra ASSIGNED",
				StatusCode: http.StatusConflict,
			}

		case errors.Is(transactionError, validations.ErrConcurrentAssignment):
			return nil, &interfaces.Error{
				Error:      "concurrent_assignment",
				Message:    "La asignación cambió durante la operación",
				StatusCode: http.StatusConflict,
			}

		default:
			pc.logging.LogError(
				"assignment_database_error",
				map[string]any{"error": transactionError.Error()},
			)

			return nil, validations.AssignmentDatabaseError()
		}
	}

	return &updatedAssignment, nil
}
