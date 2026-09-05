package recommendations

import "github.com/google/uuid"

type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Recommendation struct {
	EquipmentID               uuid.UUID `json:"equipmentId"`
	Code                      string    `json:"code"`
	Type                      string    `json:"type"`
	Brand                     string    `json:"brand"`
	Model                     string    `json:"model"`
	SerialNumber              string    `json:"serialNumber"`
	Year                      int       `json:"year"`
	CapacityTons              float64   `json:"capacityTons"`
	Location                  Location  `json:"location"`
	DistanceKM                float64   `json:"distanceKm"`
	EngineHours               float64   `json:"engineHours"`
	NextMaintenanceHours      float64   `json:"nextMaintenanceHours"`
	MaintenanceHoursRemaining float64   `json:"maintenanceHoursRemaining"`
	FuelPercent               float64   `json:"fuelPercent"`
	Score                     float64   `json:"score"`
	Reasons                   []string  `json:"reasons"`
}

type Response struct {
	RequestID       uuid.UUID        `json:"requestId"`
	Count           int              `json:"count"`
	Recommendations []Recommendation `json:"recommendations"`
}
