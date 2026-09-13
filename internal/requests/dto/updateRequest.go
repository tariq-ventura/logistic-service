package requests_dto

import "time"

type UpdateRequest struct {
	Project   *string    `json:"project" binding:"omitempty,max=250"`
	Type      *string    `json:"type" binding:"omitempty,max=100"`
	Requester *string    `json:"requester" binding:"omitempty,max=200"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}
type UpdateRequestStatus struct {
	Status string `json:"status" binding:"required,max=30"`
	Reason string `json:"reason" binding:"required,min=3,max=250"`
}
type AssignMachineryRequest struct {
	EquipmentID string `json:"equipmentId" binding:"required,uuid"`
	Reason      string `json:"reason" binding:"omitempty,max=250"`
}
