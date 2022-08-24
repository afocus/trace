package trace

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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/propagators/b3"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type exporter func(ctx context.Context) (trace.SpanExporter, error)

func ExportHTTP(endpoint string, usehttps bool) func(ctx context.Context) (trace.SpanExporter, error) {
	return func(ctx context.Context) (trace.SpanExporter, error) {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithTimeout(time.Second * 30),
		}
		if !usehttps {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)
	}
}

func ExportGRPC(endpoint string) func(ctx context.Context) (trace.SpanExporter, error) {
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

type defaultExport struct{}

func (d *defaultExport) Shutdown(ctx context.Context) error {
	return nil
}

func (d *defaultExport) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	return nil
}

func ExportStdOut() func(ctx context.Context) (trace.SpanExporter, error) {
	return func(ctx context.Context) (trace.SpanExporter, error) {
		return &defaultExport{}, nil
	}
}

func InitProvider(serviceName string, traceExporter exporter) (func(), error) {
	if traceExporter == nil {
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

	exp, err := traceExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter:%w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(trace.NewBatchSpanProcessor(exp)),
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

// InjectHttpHeader 将trace信息注入到 header
func InjectHttpHeader(ctx context.Context, header http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

// InjectMapString 将trace信息注入到 data
func InjectMapString(ctx context.Context, data map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(data))
}

// InjectMapInterface 将trace信息注入到 data
func InjectMapInterface(ctx context.Context, data map[string]interface{}) {
	otel.GetTextMapPropagator().Inject(ctx, MapInterfaceCarrier(data))
}

// ExtractHttpHeader 从http header头里提取trace信息
// 返回一个派生自ctx的具有trace信息的新context
func ExtractHttpHeader(ctx context.Context, header http.Header) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

// ExtractMapString 从map中提取trace信息
// 返回一个派生自ctx的具有trace信息的新context
func ExtractMapString(ctx context.Context, data map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(data))
}

// ExtractMapInterface 从map中提取trace信息
// 返回一个派生自ctx的具有trace信息的新context
func ExtractMapInterface(ctx context.Context, data map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, MapInterfaceCarrier(data))
}

type MapInterfaceCarrier map[string]interface{}

func (c MapInterfaceCarrier) Get(key string) string {
	value, ok := c[key]
	var valuestr string
	if ok {
		valuestr = value.(string)
	}
	return valuestr
}

func (c MapInterfaceCarrier) Set(key string, value string) {
	c[key] = value
}

func (c MapInterfaceCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
