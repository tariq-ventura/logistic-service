package requests_dto

import "time"

type LocationRequest struct {
	Name      string  `json:"name" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"gte=-90,lte=90"`
	Longitude float64 `json:"longitude" binding:"gte=-180,lte=180"`
}

type CreateRequest struct {
	EquipmentType string `json:"equipmentType" binding:"required"`
	ProjectName   string `json:"projectName" binding:"required"`

	Location LocationRequest `json:"location" binding:"required"`

	Description  string `json:"description" binding:"max=2000"`
	Requirements string `json:"requirements" binding:"max=2000"`

	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
}
