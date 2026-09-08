package requests_dto

import (
	"time"

	"github.com/google/uuid"
)

type NearLocation struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	RadiusKM  float64 `json:"radiusKm"`
}

type SearchRequestsRequest struct {
	Query            string        `json:"query,omitempty"`
	SemanticQuery    string        `json:"semanticQuery,omitempty"`
	Statuses         []string      `json:"statuses,omitempty"`
	EquipmentType    string        `json:"equipmentType,omitempty"`
	Near             *NearLocation `json:"near,omitempty"`
	MinSemanticScore *float64      `json:"minSemanticScore,omitempty"`

	Page     int `json:"page,omitempty"`
	PageSize int `json:"pageSize,omitempty"`

	QueryEmbedding    []float32 `json:"-"`
	EmbeddingProvider string    `json:"-"`
	EmbeddingModel    string    `json:"-"`
}

type SearchRequestResult struct {
	ID uuid.UUID `json:"id"`

	EquipmentType string `json:"equipmentType"`
	ProjectName   string `json:"projectName"`
	LocationName  string `json:"locationName"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Description  string `json:"description"`
	Requirements string `json:"requirements"`

	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Status    string    `json:"status"`

	DistanceKM    *float64 `json:"distanceKm,omitempty"`
	SemanticScore *float64 `json:"semanticScore,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
