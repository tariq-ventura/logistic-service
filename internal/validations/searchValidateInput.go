package validations

import (
	"net/http"

	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
)

type searchValidationError struct {
	Error      string
	Message    string
	StatusCode int
}

func ValidateSearchInput(
	input *requests_dto.SearchRequestsRequest,
) *searchValidationError {
	if len(input.Query) > 500 {
		return &searchValidationError{
			Error:      "invalid_query",
			Message:    "La consulta literal no puede superar 500 caracteres",
			StatusCode: http.StatusBadRequest,
		}
	}

	if len(input.SemanticQuery) > 500 {
		return &searchValidationError{
			Error:      "invalid_semantic_query",
			Message:    "La consulta semántica no puede superar 500 caracteres",
			StatusCode: http.StatusBadRequest,
		}
	}

	if input.MinSemanticScore != nil {
		score := *input.MinSemanticScore

		if score < 0 || score > 1 {
			return &searchValidationError{
				Error:      "invalid_semantic_score",
				Message:    "minSemanticScore debe estar entre 0 y 1",
				StatusCode: http.StatusBadRequest,
			}
		}

		if input.SemanticQuery == "" {
			return &searchValidationError{
				Error:      "semantic_query_required",
				Message:    "semanticQuery es obligatorio cuando se envía minSemanticScore",
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	if input.Near != nil {
		if input.Near.Latitude < -90 ||
			input.Near.Latitude > 90 {
			return &searchValidationError{
				Error:      "invalid_latitude",
				Message:    "La latitud debe estar entre -90 y 90",
				StatusCode: http.StatusBadRequest,
			}
		}

		if input.Near.Longitude < -180 ||
			input.Near.Longitude > 180 {
			return &searchValidationError{
				Error:      "invalid_longitude",
				Message:    "La longitud debe estar entre -180 y 180",
				StatusCode: http.StatusBadRequest,
			}
		}

		if input.Near.RadiusKM < 0 {
			return &searchValidationError{
				Error:      "invalid_radius",
				Message:    "El radio de búsqueda no puede ser negativo",
				StatusCode: http.StatusBadRequest,
			}
		}
	}

	return nil
}
