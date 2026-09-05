package requests_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) UpdateRequest(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.update_request",
		map[string]any{
			"http.method":    "PATCH",
			"http.route":     "/api/v1/requests/${id}",
			"http.params.id": id,
		},
	)
	defer span.End()

	var request requests_dto.UpdateRequest

	bindSpan, _ := rh.trace.StartSpan(ctx, "requests.update_request.BindJson", nil)
	err := c.ShouldBindJSON(&request)
	bindSpan.End()

	if err != nil {
		rh.logs.LogWarning("invalid request", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Los datos enviados no son válidos",
			"detail":  err.Error(),
		})
		return
	}

	updates := buildUpdates(request)

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "empty_update",
			"message": "Debe enviar al menos un campo",
		})
		return
	}

	dbSpan, dbCtx := rh.trace.StartSpan(ctx, "requests.database.connection", map[string]any{
		"db.name": "requests",
	})
	database := rh.db
	dbSpan.End()

	operationSpan, _ := rh.trace.StartSpan(dbCtx, "requests.database.operations", map[string]any{
		"db.name":      "requests",
		"db.operation": "update",
	})
	defer operationSpan.End()

	update, erro := database.UpdateRequest(id, updates)

	if erro != nil {
		c.JSON(erro.StatusCode, gin.H{
			"error":   erro.Error,
			"message": erro.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Peticion actualizada correctamente",
		"data":    update,
	})

}

func buildUpdates(
	request requests_dto.UpdateRequest,
) map[string]any {
	updates := make(map[string]any)

	if request.EquipmentType != nil {
		updates["equipment_type"] = strings.ToUpper(
			strings.TrimSpace(*request.EquipmentType),
		)
	}

	if request.ProjectName != nil {
		updates["project_name"] = strings.TrimSpace(
			*request.ProjectName,
		)
	}

	if request.Location != nil {
		updates["location_name"] = strings.TrimSpace(
			request.Location.Name,
		)

		updates["latitude"] = request.Location.Latitude
		updates["longitude"] = request.Location.Longitude
	}

	if request.StartDate != nil {
		updates["start_date"] = request.StartDate.UTC()
	}

	if request.EndDate != nil {
		updates["end_date"] = request.EndDate.UTC()
	}

	return updates
}
