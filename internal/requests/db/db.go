package requests_db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requests_db_postgres "github.com/tariq-ventura/logistic-service/internal/requests/db/postgres"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
)

type IRequestsDB interface {
	CreateRequest(data *requests_domain.Request, ctx context.Context) *interfaces.Error
	ListRequests(page, pageSize int, status, requestType, search string, ctx context.Context) ([]requests_domain.Request, *interfaces.Error, int64)
	ListRequestById(id uuid.UUID, ctx context.Context) (*requests_domain.Request, *interfaces.Error)
	SearchRequests(input requests_dto.SearchRequestsRequest, ctx context.Context) ([]requests_dto.SearchRequestResult, *interfaces.Error, int64)
	UpdateRequest(id uuid.UUID, updates map[string]any, ctx context.Context) (*requests_domain.Request, *interfaces.Error)
	UpdateRequestStatus(id uuid.UUID, status requests_domain.RequestStatus, reason string, ctx context.Context) (*requests_domain.Request, *requests_domain.RequestStatusHistory, *interfaces.Error)
	ListRequestStatusHistory(id uuid.UUID, ctx context.Context) ([]requests_domain.RequestStatusHistory, *interfaces.Error)
	AssignMachinery(requestID, equipmentID uuid.UUID, reason string, ctx context.Context) (*requests_domain.Request, *equipments_domain.Equipment, *requests_domain.RequestStatusHistory, *interfaces.Error)
	ReleaseMachinery(requestID uuid.UUID, reason string, ctx context.Context) (*requests_domain.Request, *interfaces.Error)
	RemoveRequest(id uuid.UUID, ctx context.Context) *interfaces.Error
}

func NewDatabase(ctx context.Context, l logging.ILogging, t interfaces.ITrace, client *gorm.DB) (IRequestsDB, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, err
	}
	if dbType != "postgresql" {
		return nil, errors.New("unsupported database backend")
	}
	return requests_db_postgres.SetupPostgres(ctx, l, t, client)
}
