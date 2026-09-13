package requests_dto

import (
	"time"

	"github.com/google/uuid"
)

type SearchRequestsRequest struct {
	Query     string   `json:"query,omitempty"`
	Statuses  []string `json:"statuses,omitempty"`
	Type      string   `json:"type,omitempty"`
	Requester string   `json:"requester,omitempty"`
	Page      int      `json:"page,omitempty"`
	PageSize  int      `json:"pageSize,omitempty"`
}
type SearchRequestResult struct {
	ID        uuid.UUID `json:"id"`
	Project   string    `json:"project"`
	Type      string    `json:"type"`
	Requester string    `json:"requester"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Status    string    `json:"status"`
	Machinery *string   `json:"machinery,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
