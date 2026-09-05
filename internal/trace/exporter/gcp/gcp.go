package trace_exporter_gcp

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const defaultEndpoint = "telemetry.googleapis.com:443"

func New(
	ctx context.Context,
	projectID string,
) (
	sdktrace.SpanExporter,
	[]attribute.KeyValue,
	error,
) {

	if projectID == "" {
		return nil, nil,
			fmt.Errorf("GCP_PROJECT_ID is required")
	}

	adc, err := oauth.NewApplicationDefault(ctx)

	if err != nil {
		return nil, nil,
			fmt.Errorf(
				"load Google application default credentials: %w",
				err,
			)
	}

	opts := []otlptracegrpc.Option{

		otlptracegrpc.WithTLSCredentials(
			credentials.NewTLS(
				&tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			),
		),

		otlptracegrpc.WithDialOption(
			grpc.WithPerRPCCredentials(adc),
		),
	}

	if os.Getenv(
		"OTEL_EXPORTER_OTLP_ENDPOINT",
	) == "" &&
		os.Getenv(
			"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		) == "" {

		opts = append(
			opts,
			otlptracegrpc.WithEndpoint(
				defaultEndpoint,
			),
		)
	}

	exporter, err :=
		otlptracegrpc.New(ctx, opts...)

	if err != nil {
		return nil, nil,
			fmt.Errorf(
				"create Google Cloud OTLP trace exporter: %w",
				err,
			)
	}

	return exporter,
		[]attribute.KeyValue{
			attribute.String(
				"gcp.project_id",
				projectID,
			),
		},
		nil
}
