package equipments_dto

type UpdateEquipmentRequest struct {
	Company        *string `json:"company" binding:"omitempty,max=150"`
	AssetNumber    *string `json:"assetNumber" binding:"omitempty,max=200"`
	Name           *string `json:"name" binding:"omitempty,max=200"`
	EquipmentClass *string `json:"equipmentClass" binding:"omitempty,max=100"`
}

type UpdateEquipmentStatusRequest struct {
	Status string `json:"status" binding:"required,max=50"`
}
