package equipments_handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	equipments_dto "github.com/tariq-ventura/logistic-service/internal/equipments/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

func (h *EquipmentHandler) CreateEquipment(c *gin.Context) {
	var input equipments_dto.CreateEquipmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	status := equipments_domain.StatusAvailable
	if strings.TrimSpace(input.Status) != "" {
		var ok bool
		status, ok = equipments_domain.NormalizeStatus(input.Status)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": "El estado de maquinaria no es válido"})
			return
		}
	}
	item := &equipments_domain.Equipment{Company: strings.TrimSpace(input.Company), AssetNumber: strings.TrimSpace(input.AssetNumber), Name: strings.TrimSpace(input.Name), EquipmentClass: strings.TrimSpace(input.EquipmentClass), Status: status}
	if err := h.db.CreateEquipment(item, c.Request.Context()); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Maquinaria registrada correctamente", "data": item})
}
func (h *EquipmentHandler) ListEquipments(c *gin.Context) {
	page := validations.ParsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := validations.ParsePositiveInt(c.DefaultQuery("pageSize", "20"), 20)
	items, err, total := h.db.ListEquipments(page, pageSize, c.Query("equipmentClass"), c.Query("status"), strings.TrimSpace(c.Query("search")), c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "pagination": gin.H{"page": page, "pageSize": pageSize, "total": total, "totalPages": validations.CalculateTotalPages(total, pageSize)}})
}
func (h *EquipmentHandler) ListEquipmentById(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "id")
	if !ok {
		return
	}
	item, err := h.db.ListEquipmentById(id, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}
func (h *EquipmentHandler) UpdateEquipment(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "id")
	if !ok {
		return
	}
	var input equipments_dto.UpdateEquipmentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	updates := make(map[string]any)
	add := func(k string, v *string) {
		if v != nil {
			updates[k] = strings.TrimSpace(*v)
		}
	}
	add("company", input.Company)
	add("asset_number", input.AssetNumber)
	add("name", input.Name)
	add("equipment_class", input.EquipmentClass)
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty_update", "message": "Debe enviar al menos un campo"})
		return
	}
	item, err := h.db.UpdateEquipment(id, updates, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Maquinaria actualizada correctamente", "data": item})
}
func (h *EquipmentHandler) UpdateEquipmentStatus(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "id")
	if !ok {
		return
	}
	var input equipments_dto.UpdateEquipmentStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Los datos enviados no son válidos", "detail": err.Error()})
		return
	}
	status, valid := equipments_domain.NormalizeStatus(input.Status)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": "El estado de maquinaria no es válido"})
		return
	}
	item, err := h.db.UpdateEquipmentStatus(id, status, c.Request.Context())
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Estado actualizado correctamente", "data": item})
}
func (h *EquipmentHandler) RemoveEquipment(c *gin.Context) {
	id, ok := validations.ParseUUIDParameter(c, "id")
	if !ok {
		return
	}
	if err := h.db.RemoveEquipment(id, c.Request.Context()); err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Error, "message": err.Message})
		return
	}
	c.Status(http.StatusNoContent)
}
