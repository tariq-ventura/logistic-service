package requests_dto

type UpdateRequestStatus struct {
	Status string `json:"status" binding:"required,oneof=PENDING ASSIGNED COMPLETED CANCELLED"`

	Reason string `json:"reason" binding:"required,min=3,max=250"`
}
