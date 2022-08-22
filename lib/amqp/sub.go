package amqp

import (
	"context"

	"github.com/afocus/trace"
	"github.com/streadway/amqp"
)

func SubHeader(name string, d *amqp.Delivery) (context.Context, error) {
	var traceID, spanID string
	for key, hh := range d.Headers {
		switch key {
		case "TraceID":
			traceID = hh.(string)
		case "SpanID":
			spanID = hh.(string)
		}
	}
	return trace.CreateFromTrace(name, traceID, spanID)
}
