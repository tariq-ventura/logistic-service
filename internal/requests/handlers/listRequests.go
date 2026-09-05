package requests_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) ListRequests(c *gin.Context) {
	ctx := c.Request.Context()

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.list_requests",
		map[string]any{
			"http.method": "GET",
			"http.route":  "/api/v1/requests",
		},
	)
	defer span.End()

	page := validations.ParsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := validations.ParsePositiveInt(c.DefaultQuery("pageSize", "20"), 20)

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

	result, err, total := database.ListRequests(page, pageSize)

	if err != nil {
		c.JSON(err.StatusCode, gin.H{
			"error":   err.Error,
			"message": err.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"pagination": gin.H{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": validations.CalculateTotalPages(total, pageSize),
		},
	})
}
