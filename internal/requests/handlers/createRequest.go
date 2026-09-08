package requests_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
)

func (rh *RequestHandler) CreateRequest(c *gin.Context) {
	ctx := c.Request.Context()
	var request requests_dto.CreateRequest

	span, _ := rh.trace.StartSpan(
		ctx,
		"equipments.create_equipment",
		map[string]any{
			"http.method": "POST",
			"http.route":  "/api/v1/requests",
		},
	)
	defer span.End()

	bindSpan, _ := rh.trace.StartSpan(ctx, "request.create_request.BindJson", nil)
	err := c.ShouldBindJSON(&request)
	bindSpan.End()

	if err != nil {
		rh.logs.LogWarning("invalid request", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Los datos enviados no son válidos",
			"detail":  err.Error(),
		})
		return
	}

	if !request.EndDate.After(request.StartDate) {
		c.JSON(
			http.StatusUnprocessableEntity,
			gin.H{
				"error": "endDate must be after startDate",
			},
		)

		return
	}

	data := requests_domain.Request{
		ID: uuid.New(),

		EquipmentType: strings.ToUpper(
			strings.TrimSpace(request.EquipmentType),
		),

		ProjectName: strings.TrimSpace(
			request.ProjectName,
		),

		LocationName: strings.TrimSpace(
			request.Location.Name,
		),

		Description:  request.Description,
		Requirements: request.Requirements,

		Latitude:  request.Location.Latitude,
		Longitude: request.Location.Longitude,

		StartDate: request.StartDate.UTC(),
		EndDate:   request.EndDate.UTC(),

		Status: requests_domain.RequestPending,
	}

	dbSpan, dbCtx := rh.trace.StartSpan(ctx, "requests.database.connection", map[string]any{
		"db.name": "requests",
	})
	database := rh.db
	dbSpan.End()

	operationSpan, _ := rh.trace.StartSpan(dbCtx, "requests.database.operations", map[string]any{
		"db.name":      "requests",
		"db.operation": "insert",
	})
	defer operationSpan.End()

	result := database.CreateRequest(data)

	if result != nil {
		c.JSON(result.StatusCode, gin.H{
			"error":   result.Error,
			"message": result.Message,
		})
		return
	}

	rh.indexRequest(ctx, &data)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Peticion registrada correctamente",
		"data":    data,
	})
}
