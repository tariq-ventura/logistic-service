package equipments_handlers

import (
	"github.com/gin-gonic/gin"
	equipments_db "github.com/tariq-ventura/logistic-service/internal/equipments/db"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
)

type EquipmentHandler struct {
	db    equipments_db.IEquipmentsDB
	trace interfaces.ITrace
	logs  logging.ILogging
}

func NewEquipmentHandler(_ *gin.Context, db equipments_db.IEquipmentsDB, trace interfaces.ITrace, logs logging.ILogging) equipments_domain.IEquipments {
	return &EquipmentHandler{db: db, trace: trace, logs: logs}
}
