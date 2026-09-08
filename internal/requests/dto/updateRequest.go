package requests_dto

import "time"

type UpdateRequest struct {
	EquipmentType *string          `json:"equipmentType"`
	ProjectName   *string          `json:"projectName"`
	Location      *LocationRequest `json:"location"`
	Description   *string          `json:"description,omitempty" binding:"omitempty,max=2000"`
	Requirements  *string          `json:"requirements,omitempty" binding:"omitempty,max=2000"`
	StartDate     *time.Time       `json:"startDate"`
	EndDate       *time.Time       `json:"endDate"`
}
