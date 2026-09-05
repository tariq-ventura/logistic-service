package requests_db_postgres

import (
	"net/http"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func (pc *PostgresClient) ListRequests(page, pageSize int) ([]requests_domain.Request, *interfaces.Error, int64) {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "equipments.database.postgres", map[string]any{
		"db.name":         "requests",
		"db.operation":    "list",
		"db.type":         "postgresql",
		"request.page":    page,
		"rquest.pageSize": pageSize,
	})
	defer operationSpan.End()

	if pageSize > 100 {
		pageSize = 100
	}

	query := pc.client.WithContext(spanCtx).Model(&requests_domain.Request{})

	var total int64

	if err := query.Count(&total).Error; err != nil {
		pc.logging.LogError("database_error", map[string]any{"error": err.Error()})
		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo contar la maquinaria",
			StatusCode: http.StatusInternalServerError,
		}, 0
	}

	var requests []requests_domain.Request

	offset := (page - 1) * pageSize

	result := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&requests)

	if result.Error != nil {
		pc.logging.LogError("database_error", map[string]any{"error": result.Error.Error()})
		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudo contar la maquinaria",
			StatusCode: http.StatusInternalServerError,
		}, 0
	}

	return requests, nil, total
}
