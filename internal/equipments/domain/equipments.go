package equipments_domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EquipmentStatus string

const (
	StatusAvailable             EquipmentStatus = "Disponible"
	StatusOccupied              EquipmentStatus = "Ocupada"
	StatusPreventiveMaintenance EquipmentStatus = "Mant. preventivo"
	StatusCorrectiveMaintenance EquipmentStatus = "Mant. correctivo"
	StatusObsolete              EquipmentStatus = "Obsoletas"
)

type Equipment struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	Company        string          `json:"company" gorm:"size:150;not null;index"`
	AssetNumber    string          `json:"assetNumber" gorm:"size:200;not null;uniqueIndex"`
	Name           string          `json:"name" gorm:"size:200;not null"`
	EquipmentClass string          `json:"equipmentClass" gorm:"size:100;not null;index"`
	Status         EquipmentStatus `json:"status" gorm:"size:50;not null;index"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt  `json:"-" gorm:"index"`
}

func (Equipment) TableName() string { return "prisma_machinery" }

func (e *Equipment) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Status == "" {
		e.Status = StatusAvailable
	}
	return nil
}

func NormalizeStatus(value string) (EquipmentStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "disponible", "available":
		return StatusAvailable, true
	case "ocupada", "ocupado", "occupied":
		return StatusOccupied, true
	case "mant. preventivo", "mantenimiento preventivo":
		return StatusPreventiveMaintenance, true
	case "mant. correctivo", "mantenimiento correctivo":
		return StatusCorrectiveMaintenance, true
	case "obsoletas", "obsoleta", "obsolete":
		return StatusObsolete, true
	default:
		return "", false
	}
}
