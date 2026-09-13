package equipments_domain

import "github.com/gin-gonic/gin"

type IEquipments interface {
	CreateEquipment(c *gin.Context)
	ListEquipments(c *gin.Context)
	ListEquipmentById(c *gin.Context)
	UpdateEquipment(c *gin.Context)
	UpdateEquipmentStatus(c *gin.Context)
	RemoveEquipment(c *gin.Context)
}
