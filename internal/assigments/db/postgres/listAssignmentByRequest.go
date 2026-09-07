package assignments_db_postgres

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
)

func (pc *PostgresClient) ListAssginmentByRequest(id uuid.UUID) (*assignments_domain.Assignment, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "assignments.database.postgres", map[string]any{
		"db.name":               "assignments",
		"db.operation":          "list",
		"db.type":               "postgresql",
		"assignments.requestId": id,
	})
	defer operationSpan.End()

	var assigment assignments_domain.Assignment

	result := pc.client.WithContext(spanCtx).First(&assigment, "request_id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, &interfaces.Error{
			Error:      "request_not_found",
			Message:    "La asignacion no existe",
			StatusCode: http.StatusNotFound,
		}
	}

	if result.Error != nil {
		pc.logging.LogError("database_error", map[string]any{"error": result.Error.Error()})
		return nil, validations.AssignmentDatabaseError()
	}

	return &assigment, nil
}
