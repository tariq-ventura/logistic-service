package requests_dto

import "time"

type CreateRequest struct {
	Project   string    `json:"project" binding:"required,max=250"`
	Type      string    `json:"type" binding:"required,max=100"`
	Requester string    `json:"requester" binding:"required,max=200"`
	StartDate time.Time `json:"startDate" binding:"required"`
	EndDate   time.Time `json:"endDate" binding:"required"`
	Status    string    `json:"status" binding:"omitempty,max=30"`
}
