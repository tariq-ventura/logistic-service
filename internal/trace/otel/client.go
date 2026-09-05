package trace_otel

import (
	"context"
	"fmt"

	"github.com/tariq-ventura/logistic-service/internal/interfaces"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Config contains backend-independent tracing metadata.
type Config struct {
	ServiceName        string
	ServiceVersion     string
	Environment        string
	ResourceAttributes []attribute.KeyValue
}

// Client implements interfaces.ITrace using the OpenTelemetry SDK.
// It does not know which backend receives the traces; that is decided by the
// sdktrace.SpanExporter injected into New.
type Client struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

type span struct {
	span trace.Span
}

func (s *span) End() {
	s.span.End()
}

func New(
	exporter sdktrace.SpanExporter,
	cfg Config,
) (*Client, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf(
			"service name is required",
		)
	}

	attrs := []attribute.KeyValue{
		attribute.String(
			"service.name",
			cfg.ServiceName,
		),
	}

	if cfg.ServiceVersion != "" {
		attrs = append(
			attrs,
			attribute.String(
				"service.version",
				cfg.ServiceVersion,
			),
		)
	}

	if cfg.Environment != "" {
		attrs = append(
			attrs,
			attribute.String(
				"deployment.environment.name",
				cfg.Environment,
			),
		)
	}

	attrs = append(
		attrs,
		cfg.ResourceAttributes...,
	)

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create trace resource: %w",
			err,
		)
	}

	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	if exporter != nil {
		providerOptions = append(
			providerOptions,
			sdktrace.WithBatcher(exporter),
		)
	} else {
		providerOptions = append(
			providerOptions,
			sdktrace.WithSampler(
				sdktrace.NeverSample(),
			),
		)
	}

	provider := sdktrace.NewTracerProvider(
		providerOptions...,
	)

	otel.SetTracerProvider(provider)

	return &Client{
		provider: provider,
		tracer: provider.Tracer(
			cfg.ServiceName,
		),
	}, nil
}

func (c *Client) StartSpan(
	ctx context.Context,
	operationName string,
	tags map[string]any,
) (interfaces.ISpan, context.Context) {
	newCtx, otelSpan := c.tracer.Start(ctx, operationName)

	attrs := make([]attribute.KeyValue, 0, len(tags))
	for key, value := range tags {
		switch v := value.(type) {
		case string:
			attrs = append(attrs, attribute.String(key, v))
		case int:
			attrs = append(attrs, attribute.Int(key, v))
		case int64:
			attrs = append(attrs, attribute.Int64(key, v))
		case bool:
			attrs = append(attrs, attribute.Bool(key, v))
		case float64:
			attrs = append(attrs, attribute.Float64(key, v))
		default:
			attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	if len(attrs) > 0 {
		otelSpan.SetAttributes(attrs...)
	}

	return &span{span: otelSpan}, newCtx
}

func (c *Client) Stop() {
	// Preserve the current interface. A future non-breaking extension can expose
	// Shutdown(context.Context) error for callers that need explicit error handling.
	_ = c.provider.Shutdown(context.Background())
}

var _ interfaces.ITrace = (*Client)(nil)
var _ interfaces.ISpan = (*span)(nil)
