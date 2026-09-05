package assignments_dto

type CreateAssignmentRequest struct {
	EquipmentID string `json:"equipmentId" binding:"required,uuid"`
	Reason      string `json:"reason" binding:"required,min=3,max=250"`
}
