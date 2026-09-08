package requests_dto

import (
	"time"

	"github.com/google/uuid"
)

type SearchRequestsRequest struct {
	Query         string        `json:"query,omitempty" binding:"omitempty,max=200"`
	Statuses      []string      `json:"statuses,omitempty" binding:"omitempty,max=10,dive,required"`
	EquipmentType string        `json:"equipmentType,omitempty" binding:"omitempty,max=30"`
	Near          *NearLocation `json:"near,omitempty"`
	Page          int           `json:"page,omitempty" binding:"omitempty,gte=1"`
	PageSize      int           `json:"pageSize,omitempty" binding:"omitempty,gte=1,lte=100"`
}

type NearLocation struct {
	Latitude  float64 `json:"latitude" binding:"gte=-90,lte=90"`
	Longitude float64 `json:"longitude" binding:"gte=-180,lte=180"`
	RadiusKM  float64 `json:"radiusKm" binding:"gt=0,lte=500"`
}

type SearchRequestResult struct {
	ID            uuid.UUID `json:"id"`
	EquipmentType string    `json:"equipmentType"`
	ProjectName   string    `json:"projectName"`
	LocationName  string    `json:"locationName"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DistanceKM    *float64  `json:"distanceKm,omitempty"`
}
