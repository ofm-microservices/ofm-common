// Package projection contains reusable migration projection orchestration.
package projection

import (
	"context"

	"github.com/ofm-microservices/ofm-common/pkg/migration/events"
)

// EventMarker atomically inserts an event marker and reports whether the event
// was new. Implementations must run the marker insert and projection in the
// same destination-database transaction.
type EventMarker interface {
	Claim(ctx context.Context, event events.Envelope) (newEvent bool, err error)
}

// Projector applies one canonical event to a legacy projection.
type Projector interface {
	Project(ctx context.Context, event events.Envelope) error
}

// Processor applies events exactly once from the projection's perspective.
// A false claim result means that the event was already processed and requires
// no projection work.
type Processor struct {
	marker    EventMarker
	projector Projector
}

// NewProcessor constructs the idempotent migration projection processor.
func NewProcessor(marker EventMarker, projector Projector) *Processor {
	return &Processor{marker: marker, projector: projector}
}

// Process claims and projects one event. If projection fails, the destination
// transaction must roll back the marker so the broker can retry the event.
func (p *Processor) Process(ctx context.Context, event events.Envelope) (bool, error) {
	newEvent, err := p.marker.Claim(ctx, event)
	if err != nil {
		return false, err
	}
	if !newEvent {
		return false, nil
	}
	if err := p.projector.Project(ctx, event); err != nil {
		return false, err
	}
	return true, nil
}
