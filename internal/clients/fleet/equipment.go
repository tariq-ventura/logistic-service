package clients_fleet

type equipmentResponse struct {
	Data Equipment `json:"data"`
}

type updateEquipmentStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
