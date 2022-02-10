package trace

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type LoggerContext struct {
	r        *zerolog.Logger
	begin    time.Time
	spanName string
	ctx      context.Context
}

func Log(c context.Context) *zerolog.Logger {
	if span := trace.SpanFromContext(c); span.SpanContext().IsValid() {
		logger := log.With().
			Str("spanID", span.SpanContext().SpanID().String()).
			Str("traceID", span.SpanContext().TraceID().String()).
			Logger()
		return &logger
	}
	return &log.Logger
}

func GetTraceID(c context.Context) string {
	if span := trace.SpanFromContext(c); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

func Attribute(key string, v interface{}) attribute.KeyValue {
	switch value := v.(type) {
	case int:
		return attribute.Int(key, value)
	case int64:
		return attribute.Int64(key, value)
	case bool:
		return attribute.Bool(key, value)
	case float32:
		return attribute.Float64(key, float64(value))
	case float64:
		return attribute.Float64(key, value)
	case fmt.Stringer:
		return attribute.Stringer(key, value)
	default:
		return attribute.String(key, fmt.Sprintf("%+v", value))
	}
}

// Start 从上下文开始一次链路追踪
// param name 轨迹的名字
// param attrs 附加属性 可以记录一些额外的信息 将会记录到链路追踪里
func Start(c context.Context, name string, attrs ...attribute.KeyValue) *LoggerContext {
	ctx, _ := defaultTracker.Start(
		c, name,
		trace.WithAttributes(attrs...),
	)
	logger := Log(ctx)
	loggerEvt := logger.Info().Str("trace", "start")
	for _, v := range attrs {
		loggerEvt.Interface(string(v.Key), v.Value.AsInterface())
	}
	loggerEvt.Msg(name)
	return &LoggerContext{
		r:        logger,
		spanName: name,
		begin:    time.Now(),
		ctx:      ctx,
	}
}

func (l *LoggerContext) WithContext(c context.Context) context.Context {
	return context.WithValue(c, traceLogContextKey{}, l)
}

func FromContext(c context.Context) *LoggerContext {
	if v := c.Value(traceLogContextKey{}); v != nil {
		lc, ok := v.(*LoggerContext)
		if ok {
			return lc
		}
	}
	return nil
}

type traceLogContextKey struct{}

func (l *LoggerContext) Log() *zerolog.Logger {
	return l.r
}

func (l *LoggerContext) Begin() time.Time {
	return l.begin
}

func (l *LoggerContext) Context() context.Context {
	return l.ctx
}

func (l *LoggerContext) Name() string {
	return l.spanName
}

func (l *LoggerContext) TraceID() string {
	return trace.SpanFromContext(l.Context()).SpanContext().TraceID().String()
}

func (l *LoggerContext) SpanID() string {
	return trace.SpanFromContext(l.Context()).SpanContext().SpanID().String()
}

func (l *LoggerContext) end(logger *zerolog.Event, status codes.Code, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(l.Context())
	logger.Dur("duration", time.Since(l.begin)).Str("trace", "end")
	for _, v := range attrs {
		logger.Interface(string(v.Key), v.Value.AsInterface())
	}
	span.SetAttributes(attrs...)
	logger.Msgf(l.spanName)
	var statusMessage string
	if err != nil {
		statusMessage = err.Error()
	}
	span.SetStatus(status, statusMessage)
	span.End()
}

// End 结束追踪，顺带附加属性
// param attrs 附加属性
func (l *LoggerContext) End(attrs ...attribute.KeyValue) {
	l.end(l.r.Info(), codes.Unset, nil, attrs...)
}

// EndOK 结束追踪并设置成功状态 附带记录一些附加属性
// param attrs 附加属性
func (l *LoggerContext) EndOK(attrs ...attribute.KeyValue) {
	l.end(l.r.Info(), codes.Ok, nil, attrs...)
}

// EndError 结束追踪并设置错误状态，顺带附加属性
// param err 错误
// param attrs 附加属性
func (l *LoggerContext) EndError(err error, attrs ...attribute.KeyValue) {
	l.end(l.r.Error().Err(err), codes.Error, err, attrs...)
}
