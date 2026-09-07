package requests_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (ah *RequestHandler) ListAssginmentByRequest(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}

	span, _ := ah.trace.StartSpan(
		ctx,
		"assignments.lists_assignments_by_request",
		map[string]any{
			"http.method":    "GET",
			"http.route":     "/api/v1/assignments/${id}",
			"http.params.id": id,
		},
	)
	defer span.End()

	dbSpan, dbCtx := ah.trace.StartSpan(ctx, "assignments.database.connection", map[string]any{
		"db.name": "assignments",
	})
	database := ah.assignments
	dbSpan.End()

	operationSpan, _ := ah.trace.StartSpan(dbCtx, "assignments.database.operations", map[string]any{
		"db.name":      "assignments",
		"db.operation": "list",
		"db.params.id": id,
	})
	defer operationSpan.End()

	result, erro := database.ListAssginmentByRequest(id)

	if erro != nil {
		c.JSON(erro.StatusCode, gin.H{
			"error":   erro.Error,
			"message": erro.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
