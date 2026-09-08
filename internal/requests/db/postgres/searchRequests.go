package requests_db_postgres

import (
	"context"
	"net/http"
	"strings"

	"github.com/pgvector/pgvector-go"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const requestDistanceSQL = `
	6371 * ACOS(
		LEAST(
			1,
			GREATEST(
				-1,
				COS(RADIANS(?))
				* COS(RADIANS(lr.latitude))
				* COS(RADIANS(lr.longitude) - RADIANS(?))
				+ SIN(RADIANS(?))
				* SIN(RADIANS(lr.latitude))
			)
		)
	)
`

func (pc *PostgresClient) SearchRequests(ctx context.Context, input requests_dto.SearchRequestsRequest) ([]requests_dto.SearchRequestResult, *interfaces.Error, int64) {
	if input.Page <= 0 {
		input.Page = 1
	}

	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	if input.PageSize > 100 {
		input.PageSize = 100
	}

	query := pc.client.
		WithContext(ctx).
		Table("logistics_requests AS lr").
		Joins(`
			LEFT JOIN logistics_request_embeddings AS lre
				ON lre.request_id = lr.id
		`)

	if value := strings.TrimSpace(input.Query); value != "" {
		pattern := "%" + value + "%"

		query = query.Where(`
			lr.project_name ILIKE ?
			OR lr.location_name ILIKE ?
			OR lr.equipment_type ILIKE ?
			OR lr.description ILIKE ?
			OR lr.requirements ILIKE ?
		`,
			pattern,
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}

	if len(input.Statuses) > 0 {
		query = query.Where(
			"lr.status IN ?",
			input.Statuses,
		)
	}

	if value := strings.TrimSpace(input.EquipmentType); value != "" {
		query = query.Where(
			"lr.equipment_type = ?",
			value,
		)
	}

	var vector pgvector.Vector
	hasSemanticQuery := len(input.QueryEmbedding) > 0

	if hasSemanticQuery {
		vector = pgvector.NewVector(input.QueryEmbedding)

		query = query.Where(
			"lre.embedding IS NOT NULL",
		)

		if input.MinSemanticScore != nil {
			query = query.Where(
				"(1 - (lre.embedding <=> ?)) >= ?",
				vector,
				*input.MinSemanticScore,
			)
		}
	}

	hasLocation := input.Near != nil

	if hasLocation {
		radiusKM := input.Near.RadiusKM
		if radiusKM <= 0 {
			radiusKM = 50
		}

		query = query.Where(
			requestDistanceSQL+" <= ?",
			input.Near.Latitude,
			input.Near.Longitude,
			input.Near.Latitude,
			radiusKM,
		)
	}

	var total int64

	countQuery := query.Session(&gorm.Session{})

	if err := countQuery.
		Distinct("lr.id").
		Count(&total).
		Error; err != nil {
		pc.logging.LogError(
			"requests_search_count_error",
			map[string]any{"error": err.Error()},
		)

		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudieron contar las solicitudes",
			StatusCode: http.StatusInternalServerError,
		}, 0
	}

	selectParts := []string{
		"lr.id",
		"lr.equipment_type",
		"lr.project_name",
		"lr.location_name",
		"lr.latitude",
		"lr.longitude",
		"lr.description",
		"lr.requirements",
		"lr.start_date",
		"lr.end_date",
		"lr.status",
		"lr.created_at",
		"lr.updated_at",
	}

	selectArguments := make([]any, 0)

	if hasSemanticQuery {
		selectParts = append(
			selectParts,
			"GREATEST(0, 1 - (lre.embedding <=> ?)) AS semantic_score",
		)

		selectArguments = append(
			selectArguments,
			vector,
		)
	} else {
		selectParts = append(
			selectParts,
			"NULL::double precision AS semantic_score",
		)
	}

	if hasLocation {
		selectParts = append(
			selectParts,
			requestDistanceSQL+" AS distance_km",
		)

		selectArguments = append(
			selectArguments,
			input.Near.Latitude,
			input.Near.Longitude,
			input.Near.Latitude,
		)
	} else {
		selectParts = append(
			selectParts,
			"NULL::double precision AS distance_km",
		)
	}

	rowsQuery := query.Select(
		strings.Join(selectParts, ",\n"),
		selectArguments...,
	)

	if hasSemanticQuery {
		rowsQuery = rowsQuery.Clauses(
			clause.OrderBy{
				Expression: clause.Expr{
					SQL:                "lre.embedding <=> ?",
					Vars:               []any{vector},
					WithoutParentheses: true,
				},
			},
		)
	}

	if hasLocation {
		rowsQuery = rowsQuery.Order(
			"distance_km ASC",
		)
	}

	rowsQuery = rowsQuery.Order(
		"lr.created_at DESC",
	)

	var requests []requests_dto.SearchRequestResult

	offset := (input.Page - 1) * input.PageSize

	result := rowsQuery.
		Limit(input.PageSize).
		Offset(offset).
		Scan(&requests)

	if result.Error != nil {
		pc.logging.LogError(
			"requests_search_error",
			map[string]any{"error": result.Error.Error()},
		)

		return nil, &interfaces.Error{
			Error:      "database_error",
			Message:    "No se pudieron buscar las solicitudes",
			StatusCode: http.StatusInternalServerError,
		}, 0
	}

	return requests, nil, total
}
