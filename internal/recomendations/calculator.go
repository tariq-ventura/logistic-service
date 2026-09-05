package recommendations

import (
	"fmt"
	"math"
	"sort"
	"strings"

	fleet_client "github.com/tariq-ventura/logistic-service/internal/clients/fleet"
)

const earthRadiusKM = 6371

func Calculate(
	requestedType string,
	projectLatitude float64,
	projectLongitude float64,
	equipments []fleet_client.Equipment,
) []Recommendation {
	result := make([]Recommendation, 0)

	for _, equipment := range equipments {
		if !strings.EqualFold(equipment.Status, "AVAILABLE") {
			continue
		}

		if !strings.EqualFold(equipment.Type, requestedType) {
			continue
		}

		maintenanceRemaining :=
			equipment.NextMaintenanceHours - equipment.EngineHours

		if maintenanceRemaining <= 0 {
			continue
		}

		distance := haversineDistance(
			projectLatitude,
			projectLongitude,
			equipment.Location.Latitude,
			equipment.Location.Longitude,
		)

		score := calculateScore(
			distance,
			maintenanceRemaining,
			equipment.FuelPercent,
		)

		result = append(result, Recommendation{
			EquipmentID:  equipment.ID,
			Code:         equipment.Code,
			Type:         equipment.Type,
			Brand:        equipment.Brand,
			Model:        equipment.Model,
			SerialNumber: equipment.SerialNumber,
			Year:         equipment.Year,
			CapacityTons: equipment.CapacityTons,
			Location: Location{
				Name:      equipment.Location.Name,
				Latitude:  equipment.Location.Latitude,
				Longitude: equipment.Location.Longitude,
			},
			DistanceKM:           round(distance, 2),
			EngineHours:          equipment.EngineHours,
			NextMaintenanceHours: equipment.NextMaintenanceHours,
			MaintenanceHoursRemaining: round(
				maintenanceRemaining,
				2,
			),
			FuelPercent: equipment.FuelPercent,
			Score:       round(score, 2),
			Reasons: []string{
				"Maquinaria disponible",
				fmt.Sprintf(
					"Se encuentra a %.2f km del proyecto",
					distance,
				),
				fmt.Sprintf(
					"Tiene %.2f horas antes del próximo mantenimiento",
					maintenanceRemaining,
				),
				fmt.Sprintf(
					"Nivel de combustible de %.2f%%",
					equipment.FuelPercent,
				),
			},
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].DistanceKM < result[j].DistanceKM
		}

		return result[i].Score > result[j].Score
	})

	return result
}

func calculateScore(
	distanceKM float64,
	maintenanceRemaining float64,
	fuelPercent float64,
) float64 {
	distanceFactor := 1 - clamp(distanceKM/200, 0, 1)
	maintenanceFactor := clamp(maintenanceRemaining/500, 0, 1)
	fuelFactor := clamp(fuelPercent/100, 0, 1)

	return (distanceFactor * 60) +
		(maintenanceFactor * 25) +
		(fuelFactor * 15)
}

func haversineDistance(
	originLatitude float64,
	originLongitude float64,
	destinationLatitude float64,
	destinationLongitude float64,
) float64 {
	originLatRadians := degreesToRadians(originLatitude)
	destinationLatRadians := degreesToRadians(destinationLatitude)

	latitudeDifference := degreesToRadians(
		destinationLatitude - originLatitude,
	)

	longitudeDifference := degreesToRadians(
		destinationLongitude - originLongitude,
	)

	a := math.Sin(latitudeDifference/2)*math.Sin(latitudeDifference/2) +
		math.Cos(originLatRadians)*
			math.Cos(destinationLatRadians)*
			math.Sin(longitudeDifference/2)*
			math.Sin(longitudeDifference/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func clamp(value float64, minimum float64, maximum float64) float64 {
	if value < minimum {
		return minimum
	}

	if value > maximum {
		return maximum
	}

	return value
}

func round(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}
