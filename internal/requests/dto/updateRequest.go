package requests_dto

import "time"

type UpdateRequest struct {
	EquipmentType *string          `json:"equipmentType"`
	ProjectName   *string          `json:"projectName"`
	Location      *LocationRequest `json:"location"`
	StartDate     *time.Time       `json:"startDate"`
	EndDate       *time.Time       `json:"endDate"`
}
