package validations

import (
	"strings"

	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
)

func NormalizeSearchInput(
	input *requests_dto.SearchRequestsRequest,
) {
	input.Query = strings.TrimSpace(input.Query)

	input.SemanticQuery = strings.TrimSpace(
		input.SemanticQuery,
	)

	input.EquipmentType = strings.ToUpper(
		strings.TrimSpace(input.EquipmentType),
	)

	for index, status := range input.Statuses {
		input.Statuses[index] = strings.ToUpper(
			strings.TrimSpace(status),
		)
	}

	if input.Page <= 0 {
		input.Page = 1
	}

	if input.PageSize <= 0 {
		input.PageSize = 20
	}

	if input.PageSize > 100 {
		input.PageSize = 100
	}
}
