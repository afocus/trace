package traceing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	tracer "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type provider func(ctx context.Context) (trace.SpanExporter, error)

var defaultTracker = otel.Tracer("github.com/afocus/traceing")

func GetDefaultTracer() tracer.Tracer {
	return defaultTracker
}

func ProviderHTTP(endpoint string) func(ctx context.Context) (trace.SpanExporter, error) {
	return func(ctx context.Context) (trace.SpanExporter, error) {
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithTimeout(time.Second*30),
		)
	}
}

func ProviderGRPC(endpoint string) func(ctx context.Context) (trace.SpanExporter, error) {
	return func(ctx context.Context) (trace.SpanExporter, error) {
		conn, err := grpc.DialContext(ctx,
			endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, err
		}
		return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	}
}

func ProviderStdout() func(ctx context.Context) (trace.SpanExporter, error) {
	return func(ctx context.Context) (trace.SpanExporter, error) {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}
}

func InitProvider(serviceName string, traceProvider provider) (func(), error) {
	if traceProvider == nil {
		return nil, errors.New("failed to create trace exporter: provider is nil")
	}
	ctx := context.Background()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource:%w", err)
	}

	traceExporter, err := traceProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter:%w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(traceExporter)),
	)
	otel.SetTracerProvider(tracerProvider)
	// 使用b3标准 这个适用于http
	otel.SetTextMapPropagator(b3.New())

	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")
	}, nil
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func InjectHttpHeader(ctx context.Context, header http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

func InjectMapString(ctx context.Context, data map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(data))
}

func ExtractHttpHeader(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

func ExtractMapString(ctx context.Context, data map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(data))
}
