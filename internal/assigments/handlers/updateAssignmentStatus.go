package assignments_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	assignments_dto "github.com/tariq-ventura/logistic-service/internal/assigments/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (ah *AssignmentHandler) UpdateAssignmentStatus(c *gin.Context) {
	ctx := c.Request.Context()

	assignmentID, ok := validations.ParseUUIDParameter(
		c,
		"assignmentID",
	)
	if !ok {
		return
	}

	var payload assignments_dto.UpdateAssignmentStatusRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Los datos enviados no son válidos",
		})
		return
	}

	newStatus := assignments_domain.AssignmentStatus(
		strings.ToUpper(strings.TrimSpace(payload.Status)),
	)

	if !newStatus.IsValidFinalStatus() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_assignment_status",
			"message": "El estado indicado no es válido",
		})
		return
	}

	assignment, databaseError :=
		ah.db.ListAssginmentById(assignmentID)

	if databaseError != nil {
		c.JSON(databaseError.StatusCode, gin.H{
			"error":   databaseError.Error,
			"message": databaseError.Message,
		})
		return
	}

	if assignment.Status != assignments_domain.AssignmentActive {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "assignment_not_active",
			"message": "La asignación ya fue finalizada",
		})
		return
	}

	equipment, fleetError := ah.fleetClient.GetEquipmentByID(
		ctx,
		assignment.EquipmentID,
	)
	if fleetError != nil {
		validations.WriteFleetError(c, fleetError)
		return
	}

	equipmentWasReleased := false

	if !strings.EqualFold(equipment.Status, "AVAILABLE") {
		fleetError = ah.fleetClient.UpdateEquipmentStatus(
			ctx,
			assignment.EquipmentID,
			"AVAILABLE",
			strings.TrimSpace(payload.Reason),
		)
		if fleetError != nil {
			validations.WriteFleetError(c, fleetError)
			return
		}

		equipmentWasReleased = true
	}

	updatedAssignment, databaseError :=
		ah.db.UpdateAssignmentStatus(
			assignmentID,
			newStatus,
			payload.Reason,
		)

	if databaseError != nil {
		if equipmentWasReleased {
			// No siempre se puede restaurar WORKING o IN_TRANSIT.
			// RESERVED evita que otro request tome el equipo.
			compensationError :=
				ah.fleetClient.UpdateEquipmentStatus(
					ctx,
					assignment.EquipmentID,
					"RESERVED",
					"Compensación por error al finalizar asignación",
				)

			if compensationError != nil {
				ah.logs.LogError(
					"assignment_compensation_failed",
					map[string]any{
						"assignmentId": assignmentID,
						"equipmentId":  assignment.EquipmentID,
						"error":        compensationError.Error(),
					},
				)
			}
		}

		c.JSON(databaseError.StatusCode, gin.H{
			"error":   databaseError.Error,
			"message": databaseError.Message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Asignación actualizada correctamente",
		"data":    updatedAssignment,
	})
}
