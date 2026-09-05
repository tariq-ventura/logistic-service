package logging

import (
	"context"
	"os"

	logging_gcp "github.com/tariq-ventura/logistic-service/internal/logging/gcp"
	logging_local "github.com/tariq-ventura/logistic-service/internal/logging/local"
)

type ILogging interface {
	LogError(msg string, fields map[string]any)
	LogWarning(msg string, fields map[string]any)
	LogInfo(msg string, fields map[string]any)
}

var NewLogging = func(ctx context.Context) (ILogging, error) {
	logging, find := os.LookupEnv("LOGGING_TYPE")

	if !find {
		logging = "local"
	}

	switch logging {
	case "GCP":
		return logging_gcp.NewGCPClient(ctx)
	default:
		return logging_local.NewLocalClient(ctx)
	}
}
