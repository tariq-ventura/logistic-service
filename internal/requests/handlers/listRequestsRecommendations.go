package requests_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	recommendations "github.com/tariq-ventura/logistic-service/internal/recomendations"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) ListRequestRecommendations(c *gin.Context) {
	ctx := c.Request.Context()

	requestID, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.list_recommendations",
		map[string]any{
			"http.method":    http.MethodGet,
			"http.route":     "/api/v1/requests/:requestID/recommendations",
			"http.requestId": requestID,
		},
	)
	defer span.End()

	request, databaseError := rh.db.ListRequestById(requestID)
	if databaseError != nil {
		c.JSON(databaseError.StatusCode, gin.H{
			"error":   databaseError.Error,
			"message": databaseError.Message,
		})
		return
	}

	if request.Status != requests_domain.RequestPending {
		c.JSON(http.StatusConflict, gin.H{
			"error": "request_not_pending",
			"message": "Solo se pueden generar recomendaciones " +
				"para peticiones con estado PENDING",
		})
		return
	}

	equipments, err := rh.fleetClient.ListAvailableEquipments(
		ctx,
		string(request.EquipmentType),
	)
	if err != nil {
		rh.logs.LogError(
			"fleet_service_error",
			map[string]any{
				"requestId": requestID,
				"error":     err.Error(),
			},
		)

		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "fleet_service_unavailable",
			"message": "No se pudo consultar la maquinaria disponible",
		})
		return
	}

	result := recommendations.Calculate(
		string(request.EquipmentType),
		request.Latitude,
		request.Longitude,
		equipments,
	)

	c.JSON(http.StatusOK, gin.H{
		"data": recommendations.Response{
			RequestID:       request.ID,
			Count:           len(result),
			Recommendations: result,
		},
	})
}
