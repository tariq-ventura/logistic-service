package assignments_db_postgres

import (
	"context"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	"gorm.io/gorm"
)

type PostgresClient struct {
	client  *gorm.DB
	logging logging.ILogging
	trace   interfaces.ITrace
	ctx     context.Context
}

var SetupPostgres = func(ctx context.Context, logs logging.ILogging, trace interfaces.ITrace, client *gorm.DB) (*PostgresClient, error) {
	return &PostgresClient{
		client:  client,
		ctx:     ctx,
		trace:   trace,
		logging: logs,
	}, nil
}
