package requests_domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Request struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Project   string         `json:"project" gorm:"size:250;not null;index"`
	Type      string         `json:"type" gorm:"size:100;not null;index"`
	Requester string         `json:"requester" gorm:"size:200;not null;index"`
	Location  string         `json:"location" gorm:"size:250;index"`
	Latitude  float64        `json:"latitude" gorm:"type:numeric(10,7)"`
	Longitude float64        `json:"longitude" gorm:"type:numeric(10,7)"`
	StartDate time.Time      `json:"startDate" gorm:"not null"`
	EndDate   time.Time      `json:"endDate" gorm:"not null"`
	Status    RequestStatus  `json:"status" gorm:"size:30;not null;index"`
	Machinery *string        `json:"machinery,omitempty" gorm:"size:200"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Request) TableName() string { return "prisma_machinery_requests" }
func (r *Request) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = RequestPending
	}
	return nil
}
