package requests_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) ListRequestStatusHistory(c *gin.Context) {
	ctx := c.Request.Context()
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}

	dbSpan, dbCtx := rh.trace.StartSpan(ctx, "requests.database.connection", map[string]any{
		"db.name": "requests",
	})
	database := rh.db
	dbSpan.End()

	operationSpan, _ := rh.trace.StartSpan(dbCtx, "requests.database.operations", map[string]any{
		"db.name":      "requests",
		"db.operation": "list",
	})
	defer operationSpan.End()

	history, dberr := database.ListRequestStatusHistory(id)

	if dberr != nil {
		c.JSON(dberr.StatusCode, gin.H{
			"error":   dberr.Error,
			"message": dberr.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
	})
}
