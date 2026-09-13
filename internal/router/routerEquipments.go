package router

import (
	"github.com/gin-gonic/gin"
	equipments_handlers "github.com/tariq-ventura/logistic-service/internal/equipments/handlers"
)

func (ro *Routes) EquipmentsRoutes(r *gin.Engine) {
	h := equipments_handlers.NewEquipmentHandler(ro.Context, ro.EquipmentsDB, ro.Trace, ro.Logging)
	routes := r.Group("/api/v1/equipments")
	{
		routes.POST("", h.CreateEquipment)
		routes.GET("", h.ListEquipments)
		routes.GET("/:id", h.ListEquipmentById)
		routes.PATCH("/:id", h.UpdateEquipment)
		routes.PATCH("/:id/status", h.UpdateEquipmentStatus)
		routes.DELETE("/:id", h.RemoveEquipment)
	}
}
