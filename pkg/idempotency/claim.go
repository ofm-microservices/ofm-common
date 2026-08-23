// Package idempotency contains durable event-consumption primitives shared by
// service projection adapters.
package idempotency

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Event identifies one canonical event for durable deduplication.
type Event struct {
	EventID          string `json:"event_id"`
	EventType        string `json:"event_type"`
	SourceService    string `json:"source_service"`
	AggregateType    string `json:"aggregate_type"`
	AggregateID      string `json:"aggregate_id"`
	AggregateVersion int64  `json:"aggregate_version"`
}

// Decode extracts event identity from the canonical envelope without coupling
// transport adapters to a service-specific event package.
func Decode(payload []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return Event{}, fmt.Errorf("decode event identity: %w", err)
	}
	if event.EventID == "" || event.EventType == "" || event.AggregateID == "" {
		return Event{}, fmt.Errorf("event identity is incomplete")
	}
	return event, nil
}

// DecodeOrFingerprint returns canonical identity when present and otherwise a
// stable identity for command payloads that predate the canonical envelope.
// The topic is part of the fallback identity so equal commands on different
// streams cannot suppress one another.
func DecodeOrFingerprint(topic string, payload []byte) Event {
	event, err := Decode(payload)
	if err == nil {
		return event
	}
	hash := sha256.Sum256(append([]byte(topic+":"), payload...))
	// processed_events uses UUID as its primary key. Keep the fallback
	// fingerprint deterministic while representing it as a valid UUID; command
	// envelopes may not contain a UUID event_id yet.
	id := fmt.Sprintf("%x-%x-%x-%x-%x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
	return Event{EventID: id, EventType: topic, SourceService: "kafka", AggregateType: topic, AggregateID: id}
}

// ClaimSQL atomically claims an event in a transaction. A false result means
// the event was already processed and the caller must skip its projection.
type Execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// Store is the durable event-claim contract used by non-SQL projections.
// Implementations must atomically report whether an event was newly claimed.
type Store interface {
	Claim(context.Context, Event) (bool, error)
	Release(context.Context, string) error
}

func ClaimSQL(ctx context.Context, tx Execer, event Event) (bool, error) {
	if tx == nil {
		return false, fmt.Errorf("transaction is nil")
	}
	result, err := tx.ExecContext(ctx, `
INSERT INTO processed_events(event_id, event_type, source_service, aggregate_type, aggregate_id, aggregate_version)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (event_id) DO NOTHING`, event.EventID, event.EventType, event.SourceService, event.AggregateType, event.AggregateID, event.AggregateVersion)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}

// ClaimDB claims an event using a standalone database operation. Callers must
// remove the claim when processing fails; successful projections retain it.
func ClaimDB(ctx context.Context, db Execer, event Event) (bool, error) {
	return ClaimSQL(ctx, db, event)
}

// Release removes a claim after a failed projection so retry can process it.
func Release(ctx context.Context, db Execer, eventID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM processed_events WHERE event_id = $1`, eventID)
	return err
}
