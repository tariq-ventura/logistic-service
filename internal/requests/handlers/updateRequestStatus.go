package requests_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (rh *RequestHandler) UpdateRequestStatus(c *gin.Context) {
	ctx := c.Request.Context()

	requestID, ok := validations.ParseUUIDParameter(c, "requestID")

	if !ok {
		return
	}

	span, _ := rh.trace.StartSpan(
		ctx,
		"requests.update_request_status",
		map[string]any{
			"http.method":    "PATCH",
			"http.route":     "/api/v1/requests/:requestID/status",
			"http.params.id": requestID,
		},
	)

	defer span.End()

	var request requests_dto.UpdateRequestStatus

	if err := c.ShouldBindJSON(&request); err != nil {
		rh.logs.LogWarning(
			"invalid_request",
			map[string]any{
				"error": err.Error(),
			},
		)

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   "invalid_request",
				"message": "Los datos enviados no son válidos",
				"detail":  err.Error(),
			},
		)

		return
	}

	newStatus := requests_domain.RequestStatus(
		strings.ToUpper(
			strings.TrimSpace(request.Status),
		),
	)

	if !newStatus.IsValid() {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error":   "invalid_status",
				"message": "El estado indicado no es válido",
			},
		)

		return
	}

	updatedRequest, createdHistory, responseError := rh.db.UpdateRequestStatus(requestID, newStatus, request.Reason)

	if responseError != nil {
		c.JSON(
			responseError.StatusCode,
			gin.H{
				"error":   responseError.Error,
				"message": responseError.Message,
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "Estado actualizado correctamente",
			"data": gin.H{
				"request":    updatedRequest,
				"transition": createdHistory,
			},
		},
	)
}
