package requests_search

import (
	"fmt"
	"strings"

	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
)

func BuildDocument(request *requests_domain.Request) string {
	parts := []string{
		fmt.Sprintf("Tipo de maquinaria: %s", request.EquipmentType),
		fmt.Sprintf("Proyecto: %s", request.ProjectName),
		fmt.Sprintf("Ubicación: %s", request.LocationName),
	}

	if value := strings.TrimSpace(request.Description); value != "" {
		parts = append(
			parts,
			fmt.Sprintf("Descripción: %s", value),
		)
	}

	if value := strings.TrimSpace(request.Requirements); value != "" {
		parts = append(
			parts,
			fmt.Sprintf("Requerimientos: %s", value),
		)
	}

	return strings.Join(parts, "\n")
}
