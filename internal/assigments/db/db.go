package assignmets_db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	assignments_db_postgres "github.com/tariq-ventura/logistic-service/internal/assigments/db/postgres"
	assignments_domain "github.com/tariq-ventura/logistic-service/internal/assigments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
)

type IAssignmentsDB interface {
	CreateAssignment(assignment assignments_domain.Assignment, reason string) *interfaces.Error
	ListAssginmentByRequest(id uuid.UUID) (*assignments_domain.Assignment, *interfaces.Error)
	ListAssginments(page, pageSize int, status string) ([]assignments_domain.Assignment, *interfaces.Error, int64)
	ListAssginmentById(id uuid.UUID) (*assignments_domain.Assignment, *interfaces.Error)
	UpdateAssignmentStatus(assignmentID uuid.UUID, newStatus assignments_domain.AssignmentStatus, reason string) (*assignments_domain.Assignment, *interfaces.Error)
}

var NewDatabase = func(ctx context.Context, l logging.ILogging, t interfaces.ITrace, client *gorm.DB) (IAssignmentsDB, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, err
	}

	switch dbType {
	case "postgresql":
		l.LogInfo("Database selected", map[string]any{"db_type": dbType})
		return assignments_db_postgres.SetupPostgres(ctx, l, t, client)
	default:
		return nil, errors.New("unsupported database backend")
	}
}
