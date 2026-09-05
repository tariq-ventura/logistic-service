package requests_db_postgres

import (
	"errors"
	"net/http"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"gorm.io/gorm"
)

func (pc *PostgresClient) CreateRequest(data requests_domain.Request) *interfaces.Error {
	operationSpan, spanCtx := pc.trace.StartSpan(pc.ctx, "equipments.database.postgres", map[string]any{
		"db.name":              "equipments",
		"db.operation":         "insert",
		"db.type":              "postgresql",
		"request.projectName":  data.ProjectName,
		"equipments.createdAt": data.CreatedAt,
	})
	defer operationSpan.End()

	result := pc.client.WithContext(spanCtx).Create(&data)

	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrDuplicatedKey):
			pc.logging.LogWarning("equipment_already_exists", nil)
			return &interfaces.Error{
				Error:      "equipment_already_exists",
				Message:    "Ya existe esta peticion",
				StatusCode: http.StatusConflict,
			}
		default:
			pc.logging.LogError("database_error", map[string]any{"error": result.Error})
			return &interfaces.Error{
				Error:      "database_error",
				Message:    "No se pudo registrar la maquinaria",
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	pc.logging.LogInfo("PostgreSQL insert success", map[string]interface{}{"insertedID": data.ID})
	return nil
}
