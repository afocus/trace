package trace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Trace struct {
	r         *zerolog.Logger
	begin     time.Time
	spanName  string
	spanAttrs []attribute.KeyValue
	span      trace.Span
}

var traceSpanPool = &sync.Pool{New: func() interface{} {
	return &Trace{}
}}

func GetLog(c context.Context) *zerolog.Logger {
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
func Start(c context.Context, name string, attrs ...attribute.KeyValue) (*Trace, context.Context) {
	ctx, span := defaultTracker.Start(c, name)
	tr := traceSpanPool.Get().(*Trace)
	tr.begin = time.Now()
	tr.r = GetLog(ctx)
	tr.spanName = name
	tr.spanAttrs = attrs
	tr.span = span
	return tr, context.WithValue(ctx, traceLogContextKey{}, tr)
}

func (l *Trace) SetAttributes(attrs ...attribute.KeyValue) {
	l.spanAttrs = append(l.spanAttrs, attrs...)
}

func FromContext(c context.Context) *Trace {
	if v := c.Value(traceLogContextKey{}); v != nil {
		lc, ok := v.(*Trace)
		if ok {
			return lc
		}
	}
	return nil
}

type traceLogContextKey struct{}

func (l *Trace) Log() *zerolog.Logger {
	if l.r == nil {
		return &log.Logger
	}
	return l.r
}

func (l *Trace) Begin() time.Time {
	return l.begin
}

func (l *Trace) Name() string {
	return l.spanName
}

func (l *Trace) TraceID() string {
	return l.span.SpanContext().TraceID().String()
}

func (l *Trace) SpanID() string {
	return l.span.SpanContext().SpanID().String()
}

func (l *Trace) End(err ...error) {
	p := l.Log().Info()
	if len(err) > 0 {
		if err[0] == nil {
			l.span.SetStatus(codes.Ok, "")
		} else {
			l.span.SetStatus(codes.Error, err[0].Error())
			p = l.Log().Error().Err(err[0])
		}
	}
	for _, v := range l.spanAttrs {
		p.Interface(string(v.Key), v.Value.AsInterface())
	}
	p.Dur("duration", time.Since(l.Begin()))
	p.Msg(l.spanName)
	l.span.SetAttributes(l.spanAttrs...)
	l.span.End()
	traceSpanPool.Put(l)
}
