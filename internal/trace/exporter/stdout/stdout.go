package trace_exporter_stdout

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func New(
	ctx context.Context,
) (
	sdktrace.SpanExporter,
	[]attribute.KeyValue,
	error,
) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)

	if err != nil {
		return nil, nil, fmt.Errorf(
			"create stdout trace exporter: %w",
			err,
		)
	}

	return exporter, nil, nil
}
