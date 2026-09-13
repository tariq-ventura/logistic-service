package main

import (
	"context"
	"log"

	"github.com/tariq-ventura/logistic-service/internal/database"
	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	"github.com/tariq-ventura/logistic-service/internal/router"
	"github.com/tariq-ventura/logistic-service/internal/trace"
)

func initTrace(ctx context.Context, l logging.ILogging) (interfaces.ITrace, error) {
	l.LogInfo("Initializing tracing system", nil)
	tracer, err := trace.NewTrace(ctx)
	if err != nil {
		l.LogError("Tracing initialization error", map[string]any{"error": err.Error()})
		return nil, err
	}
	return tracer, nil
}
func runApp(ctx context.Context, l logging.ILogging, t interfaces.ITrace) error {
	repositories, migrator, err := database.SetupDatabase(ctx, l, t)
	if err != nil {
		return err
	}
	if err := migrator.MigrateDatabase(ctx); err != nil {
		return err
	}
	routes := &router.Routes{Logging: l, Trace: t, EquipmentsDB: repositories.Equipments, RequestsDB: repositories.Requests}
	routes.Routes = routes.SetupRouter()
	routes.Run()
	return nil
}
func main() {
	ctx := context.Background()
	logs, err := logging.NewLogging(ctx)
	if err != nil {
		log.Fatalf("Logging setup error: %v", err)
	}
	tracer, err := initTrace(ctx, logs)
	if err != nil {
		return
	}
	if err := runApp(ctx, logs, tracer); err != nil {
		logs.LogError("Application setup error", map[string]any{"error": err.Error()})
	}
}
