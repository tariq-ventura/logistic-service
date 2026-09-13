package requests_db_postgres

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (pc *PostgresClient) responseError(message string, err error) *interfaces.Error {
	pc.logging.LogError("request_database_error", map[string]any{"error": err.Error()})
	return &interfaces.Error{Error: "database_error", Message: message, StatusCode: http.StatusInternalServerError}
}

func (pc *PostgresClient) CreateRequest(data *requests_domain.Request, ctx context.Context) *interfaces.Error {
	if err := pc.client.WithContext(ctx).Create(data).Error; err != nil {
		return pc.responseError("No se pudo registrar la solicitud", err)
	}
	return nil
}

func (pc *PostgresClient) ListRequests(page, pageSize int, status, requestType, search string, ctx context.Context) ([]requests_domain.Request, *interfaces.Error, int64) {
	if pageSize > 100 {
		pageSize = 100
	}
	query := pc.client.WithContext(ctx).Model(&requests_domain.Request{})
	if status != "" {
		if normalized, ok := requests_domain.NormalizeStatus(status); ok {
			query = query.Where("status = ?", normalized)
		} else {
			query = query.Where("1 = 0")
		}
	}
	if requestType != "" {
		query = query.Where("LOWER(type) = ?", strings.ToLower(strings.TrimSpace(requestType)))
	}
	if search != "" {
		pattern := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("project ILIKE ? OR type ILIKE ? OR requester ILIKE ? OR machinery ILIKE ?", pattern, pattern, pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, pc.responseError("No se pudo contar las solicitudes", err), 0
	}
	items := make([]requests_domain.Request, 0)
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		return nil, pc.responseError("No se pudieron consultar las solicitudes", err), 0
	}
	return items, nil, total
}

func (pc *PostgresClient) ListRequestById(id uuid.UUID, ctx context.Context) (*requests_domain.Request, *interfaces.Error) {
	var item requests_domain.Request
	err := pc.client.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &interfaces.Error{Error: "request_not_found", Message: "La solicitud no existe", StatusCode: http.StatusNotFound}
	}
	if err != nil {
		return nil, pc.responseError("No se pudo consultar la solicitud", err)
	}
	return &item, nil
}

func (pc *PostgresClient) UpdateRequest(id uuid.UUID, updates map[string]any, ctx context.Context) (*requests_domain.Request, *interfaces.Error) {
	item, responseError := pc.ListRequestById(id, ctx)
	if responseError != nil {
		return nil, responseError
	}
	if item.Status != requests_domain.RequestPending {
		return nil, &interfaces.Error{Error: "request_cannot_be_updated", Message: "Solo se pueden modificar solicitudes pendientes", StatusCode: http.StatusConflict}
	}
	if err := pc.client.WithContext(ctx).Model(&requests_domain.Request{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, pc.responseError("No se pudo actualizar la solicitud", err)
	}
	return pc.ListRequestById(id, ctx)
}

func (pc *PostgresClient) RemoveRequest(id uuid.UUID, ctx context.Context) *interfaces.Error {
	item, responseError := pc.ListRequestById(id, ctx)
	if responseError != nil {
		return responseError
	}
	if item.Status != requests_domain.RequestPending || item.Machinery != nil {
		return &interfaces.Error{Error: "request_cannot_be_deleted", Message: "Solo se pueden eliminar solicitudes pendientes sin maquinaria asignada", StatusCode: http.StatusConflict}
	}
	result := pc.client.WithContext(ctx).Delete(&requests_domain.Request{}, "id = ?", id)
	if result.Error != nil {
		return pc.responseError("No se pudo eliminar la solicitud", result.Error)
	}
	return nil
}

func (pc *PostgresClient) ListRequestStatusHistory(id uuid.UUID, ctx context.Context) ([]requests_domain.RequestStatusHistory, *interfaces.Error) {
	if _, err := pc.ListRequestById(id, ctx); err != nil {
		return nil, err
	}
	items := make([]requests_domain.RequestStatusHistory, 0)
	if err := pc.client.WithContext(ctx).Where("request_id = ?", id).Order("changed_at DESC").Find(&items).Error; err != nil {
		return nil, pc.responseError("No se pudo consultar el historial", err)
	}
	return items, nil
}

func (pc *PostgresClient) UpdateRequestStatus(id uuid.UUID, status requests_domain.RequestStatus, reason string, ctx context.Context) (*requests_domain.Request, *requests_domain.RequestStatusHistory, *interfaces.Error) {
	var updated requests_domain.Request
	var history requests_domain.RequestStatusHistory
	err := pc.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current requests_domain.Request
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", id).Error; err != nil {
			return err
		}
		if current.Status == status || !current.Status.CanTransitionTo(status) {
			return errors.New("invalid request status transition")
		}
		if status == requests_domain.RequestApproved && current.Machinery == nil {
			return errors.New("machinery is required before approval")
		}
		if status == requests_domain.RequestPending && current.Machinery != nil {
			if err := tx.Model(&equipments_domain.Equipment{}).Where("asset_number = ?", *current.Machinery).Update("status", equipments_domain.StatusAvailable).Error; err != nil {
				return err
			}
			current.Machinery = nil
		}
		previous := current.Status
		if err := tx.Model(&current).Updates(map[string]any{"status": status, "machinery": current.Machinery, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		history = requests_domain.RequestStatusHistory{ID: uuid.New(), RequestID: id, FromStatus: previous, ToStatus: status, Reason: strings.TrimSpace(reason), ChangedAt: time.Now().UTC()}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		return tx.First(&updated, "id = ?", id).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, &interfaces.Error{Error: "request_not_found", Message: "La solicitud no existe", StatusCode: http.StatusNotFound}
	}
	if err != nil {
		return nil, nil, &interfaces.Error{Error: "invalid_status_transition", Message: err.Error(), StatusCode: http.StatusConflict}
	}
	return &updated, &history, nil
}

func (pc *PostgresClient) AssignMachinery(requestID, equipmentID uuid.UUID, reason string, ctx context.Context) (*requests_domain.Request, *equipments_domain.Equipment, *requests_domain.RequestStatusHistory, *interfaces.Error) {
	var request requests_domain.Request
	var equipment equipments_domain.Equipment
	var history requests_domain.RequestStatusHistory
	err := pc.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&request, "id = ?", requestID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&equipment, "id = ?", equipmentID).Error; err != nil {
			return err
		}
		if request.Status != requests_domain.RequestPending {
			return errors.New("request is not pending")
		}
		if equipment.Status != equipments_domain.StatusAvailable {
			return errors.New("equipment is not available")
		}
		if !strings.EqualFold(strings.TrimSpace(request.Type), strings.TrimSpace(equipment.EquipmentClass)) {
			return errors.New("equipment class does not match request type")
		}
		previous := request.Status
		assetNumber := equipment.AssetNumber
		if err := tx.Model(&request).Updates(map[string]any{"machinery": assetNumber, "status": requests_domain.RequestApproved, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		if err := tx.Model(&equipment).Updates(map[string]any{"status": equipments_domain.StatusOccupied, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		history = requests_domain.RequestStatusHistory{ID: uuid.New(), RequestID: requestID, FromStatus: previous, ToStatus: requests_domain.RequestApproved, Reason: strings.TrimSpace(reason), ChangedAt: time.Now().UTC()}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		if err := tx.First(&request, "id = ?", requestID).Error; err != nil {
			return err
		}
		return tx.First(&equipment, "id = ?", equipmentID).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, &interfaces.Error{Error: "request_or_equipment_not_found", Message: "La solicitud o la maquinaria no existe", StatusCode: http.StatusNotFound}
	}
	if err != nil {
		return nil, nil, nil, &interfaces.Error{Error: "assignment_conflict", Message: err.Error(), StatusCode: http.StatusConflict}
	}
	return &request, &equipment, &history, nil
}

func (pc *PostgresClient) ReleaseMachinery(requestID uuid.UUID, reason string, ctx context.Context) (*requests_domain.Request, *interfaces.Error) {
	var updated requests_domain.Request
	err := pc.client.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var request requests_domain.Request
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&request, "id = ?", requestID).Error; err != nil {
			return err
		}
		if request.Machinery == nil {
			return errors.New("request has no assigned machinery")
		}
		if err := tx.Model(&equipments_domain.Equipment{}).Where("asset_number = ?", *request.Machinery).Update("status", equipments_domain.StatusAvailable).Error; err != nil {
			return err
		}
		previous := request.Status
		if err := tx.Model(&request).Updates(map[string]any{"machinery": nil, "status": requests_domain.RequestPending, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		history := requests_domain.RequestStatusHistory{ID: uuid.New(), RequestID: requestID, FromStatus: previous, ToStatus: requests_domain.RequestPending, Reason: strings.TrimSpace(reason), ChangedAt: time.Now().UTC()}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		return tx.First(&updated, "id = ?", requestID).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &interfaces.Error{Error: "request_not_found", Message: "La solicitud no existe", StatusCode: http.StatusNotFound}
	}
	if err != nil {
		return nil, &interfaces.Error{Error: "release_conflict", Message: err.Error(), StatusCode: http.StatusConflict}
	}
	return &updated, nil
}

func (pc *PostgresClient) SearchRequests(input requests_dto.SearchRequestsRequest, ctx context.Context) ([]requests_dto.SearchRequestResult, *interfaces.Error, int64) {
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}
	query := pc.client.WithContext(ctx).Model(&requests_domain.Request{})
	if input.Query != "" {
		pattern := "%" + strings.TrimSpace(input.Query) + "%"
		query = query.Where("project ILIKE ? OR type ILIKE ? OR requester ILIKE ? OR machinery ILIKE ?", pattern, pattern, pattern, pattern)
	}
	if input.Type != "" {
		query = query.Where("LOWER(type) = ?", strings.ToLower(strings.TrimSpace(input.Type)))
	}
	if input.Requester != "" {
		query = query.Where("requester ILIKE ?", "%"+strings.TrimSpace(input.Requester)+"%")
	}
	if len(input.Statuses) > 0 {
		normalized := make([]requests_domain.RequestStatus, 0, len(input.Statuses))
		for _, value := range input.Statuses {
			if status, ok := requests_domain.NormalizeStatus(value); ok {
				normalized = append(normalized, status)
			}
		}
		if len(normalized) == 0 {
			return []requests_dto.SearchRequestResult{}, nil, 0
		}
		query = query.Where("status IN ?", normalized)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, pc.responseError("No se pudieron contar las solicitudes", err), 0
	}
	items := make([]requests_dto.SearchRequestResult, 0)
	if err := query.Order("created_at DESC").Limit(input.PageSize).Offset((input.Page - 1) * input.PageSize).Scan(&items).Error; err != nil {
		return nil, pc.responseError("No se pudieron buscar las solicitudes", err), 0
	}
	return items, nil, total
}
