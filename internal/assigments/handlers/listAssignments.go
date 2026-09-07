package assignments_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (ah *AssignmentHandler) ListAssginments(c *gin.Context) {
	ctx := c.Request.Context()

	span, _ := ah.trace.StartSpan(
		ctx,
		"assignments.list_assignments",
		map[string]any{
			"http.method": "GET",
			"http.route":  "/api/v1/assignments",
		},
	)
	defer span.End()

	page := validations.ParsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := validations.ParsePositiveInt(c.DefaultQuery("pageSize", "20"), 20)
	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))

	dbSpan, dbCtx := ah.trace.StartSpan(ctx, "assignments.database.connection", map[string]any{
		"db.name": "assignments",
	})
	database := ah.db
	dbSpan.End()

	operationSpan, _ := ah.trace.StartSpan(dbCtx, "assignments.database.operations", map[string]any{
		"db.name":      "assignments",
		"db.operation": "list",
	})
	defer operationSpan.End()

	result, err, total := database.ListAssginments(page, pageSize, status)

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
