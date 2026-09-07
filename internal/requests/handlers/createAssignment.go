package requests_handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	assignments_dto "github.com/tariq-ventura/logistic-service/internal/assigments/dto"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (ah *RequestHandler) CreateAssignment(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now().UTC()
	requestID, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	var assignment assignments_dto.CreateAssignmentRequest

	span, _ := ah.trace.StartSpan(
		ctx,
		"assignments.create_assignment",
		map[string]any{
			"http.method": "POST",
			"http.route":  "/api/v1/assignments",
		},
	)
	defer span.End()

	bindSpan, _ := ah.trace.StartSpan(ctx, "assignments.create_assignment.BindJson", nil)
	err := c.ShouldBindJSON(&assignment)
	bindSpan.End()

	if err != nil {
		ah.logs.LogWarning("invalid request", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Los datos enviados no son válidos",
			"detail":  err.Error(),
		})
		return
	}

	equipmentID, err := uuid.Parse(assignment.EquipmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_equipment_id",
			"message": "El identificador de maquinaria no es válido",
		})
		return
	}

	data := assignments_domain.Assignment{
		ID:           uuid.New(),
		RequestID:    requestID,
		EquipmentID:  equipmentID,
		Status:       assignments_domain.AssignmentActive,
		StatusReason: strings.TrimSpace(assignment.Reason),
		AssignedAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	dbSpan, dbCtx := ah.trace.StartSpan(ctx, "assignments.database.connection", map[string]any{
		"db.name": "assignments",
	})
	database := ah.assignments
	dbSpan.End()

	operationSpan, _ := ah.trace.StartSpan(dbCtx, "assignments.database.operations", map[string]any{
		"db.name":      "assignments",
		"db.operation": "insert",
	})
	defer operationSpan.End()

	request, databaseError := ah.db.ListRequestById(requestID)
	if databaseError != nil {
		c.JSON(databaseError.StatusCode, gin.H{
			"error":   databaseError.Error,
			"message": databaseError.Message,
		})
		return
	}

	if request.Status != requests_domain.RequestPending {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "request_not_pending",
			"message": "La petición ya no está disponible para asignación",
		})
		return
	}

	equipment, fleetError := ah.fleetClient.GetEquipmentByID(
		ctx,
		equipmentID,
	)
	if fleetError != nil {
		validations.WriteFleetError(c, fleetError)
		return
	}

	if !strings.EqualFold(equipment.Status, "AVAILABLE") {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "equipment_not_available",
			"message": "La maquinaria ya no está disponible",
		})
		return
	}

	if !strings.EqualFold(
		equipment.Type,
		string(request.EquipmentType),
	) {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "equipment_type_mismatch",
			"message": "La maquinaria no corresponde al tipo solicitado",
		})
		return
	}

	maintenanceRemaining :=
		equipment.NextMaintenanceHours - equipment.EngineHours

	if maintenanceRemaining <= 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "equipment_maintenance_due",
			"message": "La maquinaria tiene mantenimiento vencido",
		})
		return
	}

	reason := strings.TrimSpace(assignment.Reason)

	if err := ah.fleetClient.UpdateEquipmentStatus(
		ctx,
		equipmentID,
		"RESERVED",
		reason,
	); err != nil {
		validations.WriteFleetError(c, err)
		return
	}

	create := database.CreateAssignment(
		data, reason,
	)
	if create != nil {
		compensationError := ah.fleetClient.UpdateEquipmentStatus(
			ctx,
			equipmentID,
			"AVAILABLE",
			"Compensación por error al crear la asignación",
		)

		if compensationError != nil {
			ah.logs.LogError(
				"assignment_compensation_failed",
				map[string]any{
					"requestId":   requestID,
					"equipmentId": equipmentID,
					"error":       compensationError.Error(),
				},
			)
		}

		c.JSON(create.StatusCode, gin.H{
			"error":   create.Error,
			"message": create.Message,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Maquinaria asignada correctamente",
		"data":    data,
	})

}
