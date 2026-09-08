package requests_handlers

import (
	"net/http"

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
			"invalid_search_request",
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

	validations.NormalizeSearchInput(&input)

	if validationError := validations.ValidateSearchInput(&input); validationError != nil {
		c.JSON(validationError.StatusCode, gin.H{
			"error":   validationError.Error,
			"message": validationError.Message,
		})
		return
	}

	if input.SemanticQuery != "" {
		if rh.embeddings == nil {
			rh.logs.LogError(
				"embedding_client_not_configured",
				nil,
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "embedding_service_unavailable",
				"message": "El proveedor de búsqueda semántica no está configurado",
			})
			return
		}

		embedding, err := rh.embeddings.EmbedQuery(
			ctx,
			input.SemanticQuery,
		)

		if err != nil {
			rh.logs.LogError(
				"embedding_service_error",
				map[string]any{
					"provider": rh.embeddings.Provider(),
					"model":    rh.embeddings.Model(),
					"error":    err.Error(),
				},
			)

			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "embedding_service_unavailable",
				"message": "No se pudo procesar la búsqueda semántica",
			})
			return
		}

		input.QueryEmbedding = embedding
		input.EmbeddingProvider = rh.embeddings.Provider()
		input.EmbeddingModel = rh.embeddings.Model()
	}

	result, dbError, total := rh.db.SearchRequests(
		ctx,
		input,
	)

	if dbError != nil {
		c.JSON(dbError.StatusCode, gin.H{
			"error":   dbError.Error,
			"message": dbError.Message,
		})
		return
	}

	c.Header("Accept-Query", `"application/json"`)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"count":    total,
			"requests": result,
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
