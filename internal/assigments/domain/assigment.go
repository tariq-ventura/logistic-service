package assignments_domain

import (
	"time"

	"github.com/google/uuid"
)

type AssignmentStatus string

const (
	AssignmentActive    AssignmentStatus = "ACTIVE"
	AssignmentCompleted AssignmentStatus = "COMPLETED"
	AssignmentCancelled AssignmentStatus = "CANCELLED"
)

func (s AssignmentStatus) IsValidFinalStatus() bool {
	return s == AssignmentCompleted ||
		s == AssignmentCancelled
}

type Assignment struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	RequestID   uuid.UUID `json:"requestId" gorm:"type:uuid;not null;uniqueIndex"`
	EquipmentID uuid.UUID `json:"equipmentId" gorm:"type:uuid;not null;index"`

	Status AssignmentStatus `json:"status" gorm:"size:30;not null;index"`

	StatusReason string `json:"statusReason" gorm:"size:250;not null"`

	AssignedAt  time.Time  `json:"assignedAt" gorm:"not null"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Assignment) TableName() string {
	return "assignments"
}
