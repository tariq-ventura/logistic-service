package requets_db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	requests_db_postgres "github.com/tariq-ventura/logistic-service/internal/requests/db/postgres"
	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_dto "github.com/tariq-ventura/logistic-service/internal/requests/dto"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
)

type IRequestsDB interface {
	CreateRequest(data requests_domain.Request) *interfaces.Error
	ListRequests(page, pageSize int) ([]requests_domain.Request, *interfaces.Error, int64)
	ListRequestById(id uuid.UUID) (*requests_domain.Request, *interfaces.Error)
	ListRequestStatusHistory(id uuid.UUID) ([]requests_domain.RequestStatusHistory, *interfaces.Error)
	SearchRequests(input requests_dto.SearchRequestsRequest) ([]requests_dto.SearchRequestResult, int64, *interfaces.Error)
	UpdateRequest(id uuid.UUID, updates map[string]any) (*requests_domain.Request, *interfaces.Error)
	UpdateRequestStatus(requestID uuid.UUID, newStatus requests_domain.RequestStatus, reason string) (*requests_domain.Request, *requests_domain.RequestStatusHistory, *interfaces.Error)
}

var NewDatabase = func(ctx context.Context, l logging.ILogging, t interfaces.ITrace, client *gorm.DB) (IRequestsDB, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, err
	}

	switch dbType {
	case "postgresql":
		l.LogInfo("Database selected", map[string]any{"db_type": dbType})
		return requests_db_postgres.SetupPostgres(ctx, l, t, client)
	default:
		return nil, errors.New("unsupported database backend")
	}
}
