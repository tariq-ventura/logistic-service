package requests_db_postgres

import (
	"net/http"
	"strings"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
)

const requestDistanceSQL = `
	6371 * ACOS(
		LEAST(
			1,
			GREATEST(
				-1,
				COS(RADIANS(?)) *
				COS(RADIANS(latitude)) *
				COS(RADIANS(longitude) - RADIANS(?)) +
				SIN(RADIANS(?)) *
				SIN(RADIANS(latitude))
			)
		)
	)
`

func (pc *PostgresClient) SearchRequests(input requests_dto.SearchRequestsRequest) ([]requests_dto.SearchRequestResult, int64, *interfaces.Error) {
	operationSpan, spanCtx := pc.trace.StartSpan(
		pc.ctx,
		"requests.database.postgres.search",
		map[string]any{
			"db.name":          "requests",
			"db.operation":     "search",
			"db.type":          "postgresql",
			"request.query":    input.Query,
			"request.page":     input.Page,
			"request.pageSize": input.PageSize,
		},
	)
	defer operationSpan.End()

	query := pc.client.
		WithContext(spanCtx).
		Model(&requests_domain.Request{})

	searchText := strings.TrimSpace(input.Query)

	if searchText != "" {
		searchPattern := "%" + searchText + "%"

		query = query.Where(
			`(
				project_name ILIKE ?
				OR location_name ILIKE ?
				OR equipment_type ILIKE ?
			)`,
			searchPattern,
			searchPattern,
			searchPattern,
		)
	}

	statuses := make([]string, 0, len(input.Statuses))

	for _, status := range input.Statuses {
		normalizedStatus := strings.ToUpper(strings.TrimSpace(status))

		if normalizedStatus != "" {
			statuses = append(statuses, normalizedStatus)
		}
	}

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	equipmentType := strings.ToUpper(
		strings.TrimSpace(input.EquipmentType),
	)

	if equipmentType != "" {
		query = query.Where(
			"equipment_type = ?",
			equipmentType,
		)
	}

	if input.Near != nil {
		query = query.Where(
			requestDistanceSQL+" <= ?",
			input.Near.Latitude,
			input.Near.Longitude,
			input.Near.Latitude,
			input.Near.RadiusKM,
		)
	}

	var total int64

	if err := query.Count(&total).Error; err != nil {
		pc.logging.LogError(
			"requests_search_count_error",
			map[string]any{
				"error": err.Error(),
			},
		)

		return nil, 0, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudieron contar las solicitudes",
			StatusCode: http.StatusInternalServerError,
		}
	}

	results := make(
		[]requests_dto.SearchRequestResult,
		0,
	)

	selectFields := `
	id,
	equipment_type,
	project_name,
	location_name,
	latitude,
	longitude,
	start_date,
	end_date,
	status,
	created_at,
	updated_at,
	NULL::double precision AS distance_km
`

	if input.Near != nil {
		selectFields = `
		id,
		equipment_type,
		project_name,
		location_name,
		latitude,
		longitude,
		start_date,
		end_date,
		status,
		created_at,
		updated_at,
	` + requestDistanceSQL + ` AS distance_km`

		query = query.Select(
			selectFields,
			input.Near.Latitude,
			input.Near.Longitude,
			input.Near.Latitude,
		)

		query = query.Order(
			"distance_km ASC, created_at DESC",
		)
	} else {
		query = query.
			Select(selectFields).
			Order("created_at DESC")
	}

	offset := (input.Page - 1) * input.PageSize

	if err := query.
		Limit(input.PageSize).
		Offset(offset).
		Scan(&results).
		Error; err != nil {
		pc.logging.LogError(
			"requests_search_error",
			map[string]any{
				"error": err.Error(),
			},
		)

		return nil, 0, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudieron buscar las solicitudes",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return results, total, nil
}
