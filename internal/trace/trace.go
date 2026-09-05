package trace

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	gcpexporter "github.com/tariq-ventura/logistic-service/internal/trace/exporter/gcp"
	otlpexporter "github.com/tariq-ventura/logistic-service/internal/trace/exporter/otlp"
	stdoutexporter "github.com/tariq-ventura/logistic-service/internal/trace/exporter/stdout"
	oteltrace "github.com/tariq-ventura/logistic-service/internal/trace/otel"
	"github.com/tariq-ventura/logistic-service/internal/validations"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var NewTrace = func(ctx context.Context) (interfaces.ITrace, error) {
	traceType, err := validations.RequiredEnv("TRACE_TYPE")
	if err != nil {
		return nil, err
	}

	serviceName, err := validations.RequiredEnv("SERVICE_NAME")
	if err != nil {
		return nil, err
	}

	cfg := oteltrace.Config{
		ServiceName:    serviceName,
		ServiceVersion: os.Getenv("SERVICE_VERSION"),
		Environment:    os.Getenv("ENVIRONMENT"),
	}

	var (
		exporter      sdktrace.SpanExporter
		providerAttrs []attribute.KeyValue
	)

	switch strings.ToUpper(traceType) {
	case "NONE", "DISABLED":
	case "GCP":
		projectID, err := validations.RequiredEnv("GCP_PROJECT_ID")
		if err != nil {
			return nil, err
		}

		exporter, providerAttrs, err = gcpexporter.New(ctx, projectID)
		if err != nil {
			return nil, err
		}

	case "OTLP":
		exporter, providerAttrs, err = otlpexporter.New(ctx)
		if err != nil {
			return nil, err
		}

	case "STDOUT":
		exporter, providerAttrs, err = stdoutexporter.New(ctx)

		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported trace type: %s", traceType)
	}

	cfg.ResourceAttributes = providerAttrs

	client, err := oteltrace.New(exporter, cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
