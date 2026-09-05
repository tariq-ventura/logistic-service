package requests_domain

import (
	"time"

	"github.com/google/uuid"
)

type RequestStatus string

const (
	RequestPending   RequestStatus = "PENDING"
	RequestAssigned  RequestStatus = "ASSIGNED"
	RequestCompleted RequestStatus = "COMPLETED"
	RequestCancelled RequestStatus = "CANCELLED"
)

type RequestStatusHistory struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	RequestID uuid.UUID `json:"requestId" gorm:"type:uuid;not null;index"`

	FromStatus RequestStatus `json:"fromStatus" gorm:"size:30;not null"`
	ToStatus   RequestStatus `json:"toStatus" gorm:"size:30;not null"`

	Reason string `json:"reason" gorm:"size:250;not null"`

	ChangedBy *uuid.UUID `json:"changedBy,omitempty" gorm:"type:uuid"`
	ChangedAt time.Time  `json:"changedAt" gorm:"not null"`

	Request Request `json:"-" gorm:"foreignKey:RequestID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

func (RequestStatusHistory) TableName() string {
	return "request_status_history"
}

var allowedRequestStatusTransitions = map[RequestStatus]map[RequestStatus]struct{}{
	RequestPending: {
		RequestAssigned:  {},
		RequestCancelled: {},
	},

	RequestAssigned: {
		RequestCompleted: {},
		RequestCancelled: {},
	},

	RequestCompleted: {},

	RequestCancelled: {},
}

func (s RequestStatus) IsValid() bool {
	switch s {
	case RequestPending,
		RequestAssigned,
		RequestCompleted,
		RequestCancelled:
		return true

	default:
		return false
	}
}

func (s RequestStatus) CanTransitionTo(
	next RequestStatus,
) bool {
	allowedTransitions, exists :=
		allowedRequestStatusTransitions[s]

	if !exists {
		return false
	}

	_, allowed := allowedTransitions[next]

	return allowed
}
