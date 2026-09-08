package requests_domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Request struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	EquipmentType string `json:"equipmentType" gorm:"size:30;not null;index"`
	ProjectName   string `json:"projectName" gorm:"size:150;not null"`
	LocationName  string `json:"locationName" gorm:"size:200;not null"`

	Latitude  float64 `json:"latitude" gorm:"not null"`
	Longitude float64 `json:"longitude" gorm:"not null"`

	Description  string `json:"description" gorm:"type:text;not null;default:''"`
	Requirements string `json:"requirements" gorm:"type:text;not null;default:''"`

	StartDate time.Time `json:"startDate" gorm:"not null"`
	EndDate   time.Time `json:"endDate" gorm:"not null"`

	Status RequestStatus `json:"status" gorm:"size:30;not null;index"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Request) TableName() string {
	return "logistics_requests"
}

func (r *Request) BeforeCreate(
	tx *gorm.DB,
) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	if r.Status == "" {
		r.Status = RequestPending
	}

	return nil
}

func (r Request) CanTransitionTo(
	next RequestStatus,
) bool {
	switch r.Status {
	case RequestPending:
		return next == RequestAssigned ||
			next == RequestCancelled

	case RequestAssigned:
		return next == RequestCompleted ||
			next == RequestCancelled

	default:
		return false
	}
}

func IsValidStatus(status RequestStatus) bool {
	switch status {
	case RequestPending,
		RequestAssigned,
		RequestCompleted,
		RequestCancelled:
		return true

	default:
		return false
	}
}
