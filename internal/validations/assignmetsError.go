package validations

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	clients_fleet "github.com/tariq-ventura/logistic-service/internal/clients/fleet"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
)

var (
	ErrRequestNotPending = errors.New(
		"request is not pending",
	)

	ErrRequestNotAssigned = errors.New(
		"request is not assigned",
	)

	ErrAssignmentNotActive = errors.New(
		"assignment is not active",
	)

	ErrConcurrentAssignment = errors.New(
		"assignment changed concurrently",
	)
)

func WriteFleetError(c *gin.Context, err error) {
	var upstreamError *clients_fleet.HTTPError

	if errors.As(err, &upstreamError) {
		switch upstreamError.StatusCode {
		case http.StatusNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "equipment_not_found",
				"message": "La maquinaria no existe",
			})

		case http.StatusConflict, http.StatusUnprocessableEntity:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "equipment_not_available",
				"message": "La maquinaria ya no está disponible",
			})

		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "fleet_service_error",
				"message": "Fleet no pudo procesar la operación",
			})
		}

		return
	}

	c.JSON(http.StatusBadGateway, gin.H{
		"error":   "fleet_service_unavailable",
		"message": "No se pudo establecer comunicación con Fleet",
	})
}

func AssignmentNotFound() *interfaces.Error {
	return &interfaces.Error{
		Error:      "assignment_not_found",
		Message:    "La asignación no existe",
		StatusCode: http.StatusNotFound,
	}
}

func AssignmentDatabaseError() *interfaces.Error {
	return &interfaces.Error{
		Error:      "database_error",
		Message:    "No se pudo consultar la asignación",
		StatusCode: http.StatusInternalServerError,
	}
}
