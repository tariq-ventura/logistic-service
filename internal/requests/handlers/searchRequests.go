package requests_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) SearchRequests(c *gin.Context) {
	ctx := c.Request.Context()

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.search",
		map[string]any{
			"http.method": "QUERY",
			"http.route":  "/api/v1/requests",
		},
	)
	defer span.End()

	var input requests_dto.SearchRequestsRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		rh.logs.LogWarning(
			"invalid search request",
			map[string]any{
				"error": err.Error(),
			},
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Los criterios de búsqueda no son válidos",
			"detail":  err.Error(),
		})
		return
	}

	input.Query = strings.TrimSpace(input.Query)
	input.EquipmentType = strings.ToUpper(
		strings.TrimSpace(input.EquipmentType),
	)

	if input.Page == 0 {
		input.Page = 1
	}

	if input.PageSize == 0 {
		input.PageSize = 20
	}

	results, total, databaseError := rh.db.SearchRequests(input)

	if databaseError != nil {
		c.JSON(databaseError.StatusCode, gin.H{
			"error":   databaseError.Error,
			"message": databaseError.Message,
		})
		return
	}

	// Indica que este recurso acepta QUERY con JSON.
	c.Header("Accept-Query", `"application/json"`)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"count":    total,
			"requests": results,
		},
		"pagination": gin.H{
			"page":     input.Page,
			"pageSize": input.PageSize,
			"total":    total,
			"totalPages": validations.CalculateTotalPages(
				total,
				input.PageSize,
			),
		},
	})
}
