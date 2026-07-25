package natstrace

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type carrier struct {
	hdr nats.Header
}

func (c carrier) Get(key string) string {
	return c.hdr.Get(key)
}

func (c carrier) Set(key, value string) {
	c.hdr.Set(key, value)
}

func (c carrier) Keys() []string {
	keys := make([]string, 0, len(c.hdr))
	for k := range c.hdr {
		keys = append(keys, k)
	}
	return keys
}

// NewMessage creates a NATS message with trace context injected into headers.
func NewMessage(ctx context.Context, subject string, payload []byte) *nats.Msg {
	msg := nats.NewMsg(subject)
	msg.Data = payload
	if msg.Header == nil {
		msg.Header = nats.Header{}
	}
	otel.GetTextMapPropagator().Inject(ctx, carrier{hdr: msg.Header})
	return msg
}

// ContextFromMessage extracts trace context from a NATS message.
func ContextFromMessage(ctx context.Context, msg *nats.Msg) context.Context {
	if msg == nil {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if len(msg.Header) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier{hdr: msg.Header})
}

// Propagator returns the propagator used for NATS trace propagation.
func Propagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
