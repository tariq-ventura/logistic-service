package clients_fleet

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Equipment struct {
	ID                   uuid.UUID  `json:"id"`
	Code                 string     `json:"code"`
	FleetID              *uuid.UUID `json:"fleetId,omitempty"`
	Type                 string     `json:"type"`
	Brand                string     `json:"brand"`
	Model                string     `json:"model"`
	SerialNumber         string     `json:"serialNumber"`
	Year                 int        `json:"year"`
	CapacityTons         float64    `json:"capacityTons"`
	Status               string     `json:"status"`
	Location             Location   `json:"location"`
	EngineHours          float64    `json:"engineHours"`
	NextMaintenanceHours float64    `json:"nextMaintenanceHours"`
	FuelPercent          float64    `json:"fuelPercent"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ListEquipmentsResponse struct {
	Data       []Equipment `json:"data"`
	Pagination Pagination  `json:"pagination"`
}
