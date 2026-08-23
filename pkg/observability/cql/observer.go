package cql

import (
	"context"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Observer converts completed gocql operations into database spans. The
// driver supplies the original query context and precise start/end times.
type Observer struct{ Service string }

// ObserveQuery records one CQL statement without attaching statement values
// to telemetry, preventing secrets and high-cardinality data leakage.
func (o Observer) ObserveQuery(ctx context.Context, q gocql.ObservedQuery) {
	service := o.Service
	if service == "" {
		service = "ofm"
	}
	_, span := otel.Tracer(service+"/database").Start(ctx, "scylla.query", trace.WithTimestamp(q.Start), trace.WithAttributes(
		attribute.String("db.system", "cassandra"),
		attribute.String("db.operation.name", "cql"),
		attribute.String("db.namespace", q.Keyspace),
		attribute.Int("db.cql.rows", q.Rows),
	))
	span.SetAttributes(attribute.Int64("db.query.duration_ms", q.End.Sub(q.Start).Milliseconds()))
	if q.Err != nil {
		span.RecordError(q.Err)
		span.SetStatus(codes.Error, q.Err.Error())
	}
	span.End(trace.WithTimestamp(q.End))
}
