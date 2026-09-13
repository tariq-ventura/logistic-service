package equipments_db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	equipments_db_postgres "github.com/tariq-ventura/logistic-service/internal/equipments/db/postgres"
	equipments_domain "github.com/tariq-ventura/logistic-service/internal/equipments/domain"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"gorm.io/gorm"
)

type IEquipmentsDB interface {
	CreateEquipment(data *equipments_domain.Equipment, ctx context.Context) *interfaces.Error
	ListEquipments(page, pageSize int, equipmentClass, status, search string, ctx context.Context) ([]equipments_domain.Equipment, *interfaces.Error, int64)
	ListEquipmentById(id uuid.UUID, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error)
	UpdateEquipment(id uuid.UUID, updates map[string]any, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error)
	UpdateEquipmentStatus(id uuid.UUID, status equipments_domain.EquipmentStatus, ctx context.Context) (*equipments_domain.Equipment, *interfaces.Error)
	RemoveEquipment(id uuid.UUID, ctx context.Context) *interfaces.Error
}

func NewDatabase(ctx context.Context, l logging.ILogging, t interfaces.ITrace, client *gorm.DB) (IEquipmentsDB, error) {
	dbType, err := validations.RequiredEnv("DB_CONTEXT")
	if err != nil {
		return nil, err
	}
	if dbType != "postgresql" {
		return nil, errors.New("unsupported database backend")
	}
	return equipments_db_postgres.SetupPostgres(ctx, l, t, client)
}
