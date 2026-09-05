package requests_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) ListRequestById(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.lists_requests_by_id",
		map[string]any{
			"http.method":    "GET",
			"http.route":     "/api/v1/requests/${id}",
			"http.params.id": id,
		},
	)
	defer span.End()

	dbSpan, dbCtx := rh.trace.StartSpan(ctx, "requests.database.connection", map[string]any{
		"db.name": "requests",
	})
	database := rh.db
	dbSpan.End()

	operationSpan, _ := rh.trace.StartSpan(dbCtx, "fleets.database.operations", map[string]any{
		"db.name":      "fleets",
		"db.operation": "list",
		"db.params.id": id,
	})
	defer operationSpan.End()

	result, erro := database.ListRequestById(id)

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
