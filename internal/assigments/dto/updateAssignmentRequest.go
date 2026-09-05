package assignments_dto

type UpdateAssignmentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=COMPLETED CANCELLED"`
	Reason string `json:"reason" binding:"required,min=3,max=250"`
}
