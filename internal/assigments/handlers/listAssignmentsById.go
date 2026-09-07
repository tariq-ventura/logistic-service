package assignments_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (ah *AssignmentHandler) ListAssginmentById(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := validations.ParseUUIDParameter(c, "assignmentID")
	if !ok {
		return
	}

	span, _ := ah.trace.StartSpan(
		ctx,
		"assignments.lists_assignments_by_id",
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
	database := ah.db
	dbSpan.End()

	operationSpan, _ := ah.trace.StartSpan(dbCtx, "assignments.database.operations", map[string]any{
		"db.name":      "assignments",
		"db.operation": "list",
		"db.params.id": id,
	})
	defer operationSpan.End()

	result, erro := database.ListAssginmentById(id)

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
