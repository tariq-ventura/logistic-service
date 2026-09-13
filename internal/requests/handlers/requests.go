package requests_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (h *RequestHandler) CreateRequest(c *gin.Context) {
	var input requests_dto.CreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	if !input.EndDate.After(input.StartDate) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_period", "message": "endDate debe ser posterior a startDate"})
		return
	}
	status := requests_domain.RequestPending
	if input.Status != "" {
		var ok bool
		status, ok = requests_domain.NormalizeStatus(input.Status)
		if !ok || status != requests_domain.RequestPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": "Una solicitud nueva debe iniciar Pendiente"})
			return
		}
	}
	item := &requests_domain.Request{Project: strings.TrimSpace(input.Project), Type: strings.TrimSpace(input.Type), Requester: strings.TrimSpace(input.Requester), StartDate: input.StartDate.UTC(), EndDate: input.EndDate.UTC(), Status: status}
	if err := h.db.CreateRequest(item, c.Request.Context()); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Solicitud registrada correctamente", "data": item})
}

func (h *RequestHandler) ListRequests(c *gin.Context) {
	page := validations.ParsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := validations.ParsePositiveInt(c.DefaultQuery("pageSize", "20"), 20)
	items, err, total := h.db.ListRequests(page, pageSize, c.Query("status"), c.Query("type"), strings.TrimSpace(c.Query("search")), c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "pagination": gin.H{"page": page, "pageSize": pageSize, "total": total, "totalPages": validations.CalculateTotalPages(total, pageSize)}})
}

func (h *RequestHandler) ListRequestById(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	item, err := h.db.ListRequestById(id, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *RequestHandler) UpdateRequest(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	var input requests_dto.UpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	current, responseError := h.db.ListRequestById(id, c.Request.Context())
	if responseError != nil {
		c.JSON(responseError.StatusCode, gin.H{"error": responseError.Error, "message": responseError.Message})
		return
	}
	startDate := current.StartDate
	endDate := current.EndDate
	updates := make(map[string]any)
	if input.Project != nil {
		updates["project"] = strings.TrimSpace(*input.Project)
	}
	if input.Type != nil {
		updates["type"] = strings.TrimSpace(*input.Type)
	}
	if input.Requester != nil {
		updates["requester"] = strings.TrimSpace(*input.Requester)
	}
	if input.StartDate != nil {
		startDate = input.StartDate.UTC()
		updates["start_date"] = startDate
	}
	if input.EndDate != nil {
		endDate = input.EndDate.UTC()
		updates["end_date"] = endDate
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty_update", "message": "Debe enviar al menos un campo"})
		return
	}
	if !endDate.After(startDate) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_period", "message": "endDate debe ser posterior a startDate"})
		return
	}
	item, err := h.db.UpdateRequest(id, updates, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Solicitud actualizada correctamente", "data": item})
}

func (h *RequestHandler) UpdateRequestStatus(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	var input requests_dto.UpdateRequestStatus
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	status, valid := requests_domain.NormalizeStatus(input.Status)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": "El estado de solicitud no es válido"})
		return
	}
	item, history, err := h.db.UpdateRequestStatus(id, status, strings.TrimSpace(input.Reason), c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Estado actualizado correctamente", "data": gin.H{"request": item, "transition": history}})
}

func (h *RequestHandler) ListRequestStatusHistory(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	items, err := h.db.ListRequestStatusHistory(id, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *RequestHandler) AssignMachinery(c *gin.Context) {
	requestID, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	var input requests_dto.AssignMachineryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	equipmentID, err := uuid.Parse(input.EquipmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_equipment_id", "message": "equipmentId debe ser un UUID válido"})
		return
	}
	request, equipment, history, responseError := h.db.AssignMachinery(requestID, equipmentID, strings.TrimSpace(input.Reason), c.Request.Context())
	if responseError != nil {
		c.JSON(responseError.StatusCode, gin.H{"error": responseError.Error, "message": responseError.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Maquinaria asignada y solicitud aprobada correctamente", "data": gin.H{"request": request, "equipment": equipment, "transition": history}})
}

func (h *RequestHandler) ReleaseMachinery(c *gin.Context) {
	requestID, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	reason := strings.TrimSpace(c.Query("reason"))
	if reason == "" {
		reason = "Liberación solicitada"
	}
	request, err := h.db.ReleaseMachinery(requestID, reason, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Maquinaria liberada correctamente", "data": request})
}

func (h *RequestHandler) RemoveRequest(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "requestID")
	if !ok {
		return
	}
	if err := h.db.RemoveRequest(id, c.Request.Context()); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RequestHandler) SearchRequests(c *gin.Context) {
	var input requests_dto.SearchRequestsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los criterios de búsqueda no son válidos", "detail": err.Error()})
		return
	}
	input.Query = strings.TrimSpace(input.Query)
	input.Type = strings.TrimSpace(input.Type)
	input.Requester = strings.TrimSpace(input.Requester)
	if len(input.Query) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_query", "message": "La consulta no puede superar 500 caracteres"})
		return
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	items, err, total := h.db.SearchRequests(input, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"count": total, "requests": items}, "pagination": gin.H{"page": input.Page, "pageSize": input.PageSize, "total": total, "totalPages": validations.CalculateTotalPages(total, input.PageSize)}})
}
