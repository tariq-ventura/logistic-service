package trace_exporter_otlp

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func New(ctx context.Context) (sdktrace.SpanExporter, []attribute.KeyValue, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	return exporter, nil, nil
}
