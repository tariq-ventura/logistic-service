package equipments_db_postgres

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"gorm.io/gorm"
)

func (pc *PostgresClient) responseError(message string, err error) *interfaces.Error {
	pc.logging.LogError("equipment_database_error", map[string]any{"error": err.Error()})
	return &interfaces.Error{Error: "database_error", Message: message, StatusCode: http.StatusInternalServerError}
}
func (pc *PostgresClient) CreateEquipment(data *equipments_domain.Equipment, ctx context.Context) *interfaces.Error {
	result := pc.client.WithContext(ctx).Create(data)
	if result.Error == nil {
		return nil
	}
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return &interfaces.Error{Error: "equipment_already_exists", Message: "Ya existe una maquinaria con ese número de activo", StatusCode: http.StatusConflict}
	}
	return pc.responseError("No se pudo registrar la maquinaria", result.Error)
}
func (pc *PostgresClient) ListEquipments(page, pageSize int, equipmentClass, status, search string, ctx context.Context) ([]equipments_domain.Equipment, *interfaces.Error, int64) {
	if pageSize > 100 {
		pageSize = 100
	}
	query := pc.client.WithContext(ctx).Model(&equipments_domain.Equipment{})
	if equipmentClass != "" {
		query = query.Where("LOWER(equipment_class) = ?", strings.ToLower(strings.TrimSpace(equipmentClass)))
	}
	if status != "" {
		query = query.Where("LOWER(status) = ?", strings.ToLower(strings.TrimSpace(status)))
	}
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("company ILIKE ? OR asset_number ILIKE ? OR name ILIKE ? OR equipment_class ILIKE ?", pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, pc.responseError("No se pudo contar la maquinaria", err), 0
	}
	items := make([]equipments_domain.Equipment, 0)
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		return nil, pc.responseError("No se pudo consultar la maquinaria", err), 0
	}
	return items, nil, total
}
func (pc *PostgresClient) ListEquipmentById(id uuid.UUID, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error) {
	var item equipments_domain.Equipment
	err := pc.client.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &interfaces.Error{Error: "equipment_not_found", Message: "La maquinaria no existe", StatusCode: http.StatusNotFound}
	}
	if err != nil {
		return nil, pc.responseError("No se pudo consultar la maquinaria", err)
	}
	return &item, nil
}
func (pc *PostgresClient) UpdateEquipment(id uuid.UUID, updates map[string]any, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error) {
	if _, err := pc.ListEquipmentById(id, ctx); err != nil {
		return nil, err
	}
	result := pc.client.WithContext(ctx).Model(&equipments_domain.Equipment{}).Where("id = ?", id).Updates(updates)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, &interfaces.Error{Error: "equipment_already_exists", Message: "Ya existe una maquinaria con ese número de activo", StatusCode: http.StatusConflict}
	}
	if result.Error != nil {
		return nil, pc.responseError("No se pudo actualizar la maquinaria", result.Error)
	}
	return pc.ListEquipmentById(id, ctx)
}
func (pc *PostgresClient) UpdateEquipmentStatus(id uuid.UUID, status equipments_domain.EquipmentStatus, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error) {
	return pc.UpdateEquipment(id, map[string]any{"status": status}, ctx)
}
func (pc *PostgresClient) RemoveEquipment(id uuid.UUID, ctx context.Context) *interfaces.Error {
	result := pc.client.WithContext(ctx).Delete(&equipments_domain.Equipment{}, "id = ?", id)
	if result.Error != nil {
		return pc.responseError("No se pudo eliminar la maquinaria", result.Error)
	}
	if result.RowsAffected == 0 {
		return &interfaces.Error{Error: "equipment_not_found", Message: "La maquinaria no existe", StatusCode: http.StatusNotFound}
	}
	return nil
}
