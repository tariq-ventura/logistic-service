package equipments_dto

type CreateEquipmentRequest struct {
	Company        string `json:"company" binding:"required,max=150"`
	AssetNumber    string `json:"assetNumber" binding:"required,max=200"`
	Name           string `json:"name" binding:"required,max=200"`
	EquipmentClass string `json:"equipmentClass" binding:"required,max=100"`
	Status         string `json:"status" binding:"omitempty,max=50"`
}
