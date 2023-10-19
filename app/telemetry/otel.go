package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
	"go.opentelemetry.io/otel/semconv/v1.20.0/netconv"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	"github.com/chatbot-go/app/config"
	"github.com/chatbot-go/app/library/ctxkey"
)

var (
	otelEnv     = "undefined"
	otelVersion = "undefined"
)

type otelCtxKey struct{}

type Otel struct {
	Tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	exporter *otlptrace.Exporter
}

func (t *Otel) Close(ctx context.Context) error {
	const operation = "Telemetry.Otel.Close"

	var errs error

	if err := t.provider.ForceFlush(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> provider flush: %w", operation, err))
	}

	if err := t.provider.Shutdown(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> provider shutdown: %w", operation, err))
	}

	if err := t.exporter.Shutdown(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> exporter shutdown: %w", operation, err))
	}

	return errs
}

func NewOtel(ctx context.Context, cfg config.Otel, env, version string) (Otel, error) {
	const operation = "Telemetry.NewOtel"

	otelEnv, otelVersion = env, version

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.ExporterTimeout),
	))
	if err != nil {
		return Otel{}, fmt.Errorf("%s -> new exporter: %w", operation, err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithExportTimeout(cfg.ExporterTimeout)),
		sdktrace.WithResource(otelResource(cfg, env, version)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
	)

	tracer := provider.Tracer(cfg.ServiceName)

	// Set global tracer provider & text propagators
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader)),
		propagation.TraceContext{},
		propagation.Baggage{},
		xray.Propagator{},
	))

	return Otel{
		tracer,
		provider,
		exporter,
	}, nil
}

// ContextWithTracer returns a new context derived from ctx that
// is associated with the given tracer.
func ContextWithTracer(parent context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(parent, otelCtxKey{}, tracer)
}

// StartInternalSpan starts a new Span with kind trace.SpanKindInternal.
func StartInternalSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tracerFromCtx(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
}

// StartServerSpan starts a new Span with kind trace.SpanKindServer.
func StartServerSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tracerFromCtx(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
}

// StartClientSpan starts a new Span with kind trace.SpanKindClient.
func StartClientSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tracerFromCtx(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindClient))
}

// StartProducerSpan starts a new Span with kind trace.SpanKindProducer.
func StartProducerSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tracerFromCtx(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
}

// StartConsumerSpan starts a new Span with kind trace.SpanKindConsumer.
func StartConsumerSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return tracerFromCtx(ctx).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
}

// AttributesFromContext generates a list of otel attributes from context keys.
func AttributesFromContext(ctx context.Context) []attribute.KeyValue {
	// https://docs.datadoghq.com/tracing/trace_collection/tracing_naming_convention/#span-tag-naming-convention
	attrs := []attribute.KeyValue{
		attribute.String("env", otelEnv),
		attribute.String("version", otelVersion),
		attribute.String("language", "go"),
		attribute.String("component", "telemetry.otel"),
	}

	if requestID, ok := ctxkey.GetRequestID(ctx); ok {
		attrs = append(attrs, attribute.String("request_id", requestID))
	}

	return attrs
}

// AttributesFromRequest generates a list of server otel attributes from http request.
func AttributesFromRequest(req *http.Request) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0)

	attrs = append(attrs, semconv.HTTPTarget(req.URL.String()))

	attrs = append(attrs, httpconv.ServerRequest("", req)...)
	attrs = append(attrs, httpconv.RequestHeader(req.Header)...)

	attrs = append(attrs, netconv.Server(req.Host, nil)...)
	attrs = append(attrs, netconv.Transport("tcp"))

	return attrs
}

func otelResource(cfg config.Otel, env, version string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.DeploymentEnvironmentKey.String(env),
		semconv.ServiceInstanceIDKey.String(uuid.NewString()),
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceNamespaceKey.String(cfg.ServiceNamespace),
		semconv.ServiceVersionKey.String(version),
		attribute.String("library.language", "go"),
	)
}

func tracerFromCtx(ctx context.Context) trace.Tracer {
	tracer, ok := ctx.Value(otelCtxKey{}).(trace.Tracer)
	if !ok {
		return nil
	}

	return tracer
}
