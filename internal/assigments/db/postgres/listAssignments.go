package assignments_db_postgres

import (
	"net/http"

	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (pc *PostgresClient) ListAssginments(page, pageSize int, status string) ([]assignments_domain.Assignment, *interfaces.Error, int64) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "assignmets.database.postgres", map[string]any{
		"db.name":         "assignmets",
		"db.operation":    "list",
		"db.type":         "postgresql",
		"request.page":    page,
		"rquest.pageSize": pageSize,
	})
	defer operationSpan.End()

	if pageSize > 100 {
		pageSize = 100
	}

	query := pc.client.WithContext(spanCtx).Model(&assignments_domain.Assignment{})

	var total int64

	if err := query.Count(&total).Error; err != nil {
		pc.logging.LogError("database_error", map[string]any{"error": err.Error()})
		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo contar las assignaciones",
			StatusCode: http.StatusInternalServerError,
		}, 0
	}

	var assigments []assignments_domain.Assignment

	offset := (page - 1) * pageSize

	result := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&assigments)

	if result.Error != nil {
		pc.logging.LogError("database_error", map[string]any{"error": result.Error.Error()})
		return nil, validations.AssignmentDatabaseError(), 0
	}

	return assigments, nil, total
}
