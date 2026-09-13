package requests_domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	RequestPending  RequestStatus = "Pendiente"
	RequestApproved RequestStatus = "Aprobada"
)

func NormalizeStatus(value string) (RequestStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "pendiente", "pending":
		return RequestPending, true
	case "aprobada", "aprobado", "approved":
		return RequestApproved, true
	default:
		return "", false
	}
}
func (s RequestStatus) CanTransitionTo(next RequestStatus) bool {
	return (s == RequestPending && next == RequestApproved) || (s == RequestApproved && next == RequestPending)
}

type RequestStatusHistory struct {
	ID         uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey"`
	RequestID  uuid.UUID     `json:"requestId" gorm:"type:uuid;not null;index"`
	FromStatus RequestStatus `json:"fromStatus" gorm:"size:30;not null"`
	ToStatus   RequestStatus `json:"toStatus" gorm:"size:30;not null"`
	Reason     string        `json:"reason" gorm:"size:250;not null"`
	ChangedAt  time.Time     `json:"changedAt" gorm:"not null"`
	Request    Request       `json:"-" gorm:"foreignKey:RequestID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (RequestStatusHistory) TableName() string { return "prisma_request_status_history" }
