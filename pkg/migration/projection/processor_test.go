package projection

import (
	"context"
	"errors"
	"testing"

	"github.com/ofm-microservices/ofm-common/pkg/migration/events"
)

type markerStub struct{ newEvent bool }

func (m markerStub) Claim(context.Context, events.Envelope) (bool, error) { return m.newEvent, nil }

type projectorStub struct {
	calls int
	err   error
}

func (p *projectorStub) Project(context.Context, events.Envelope) error { p.calls++; return p.err }

func TestProcessorSkipsDuplicate(t *testing.T) {
	projector := &projectorStub{}
	processed, err := NewProcessor(markerStub{newEvent: false}, projector).Process(context.Background(), events.Envelope{})
	if err != nil || processed || projector.calls != 0 {
		t.Fatalf("duplicate should be skipped: processed=%v err=%v calls=%d", processed, err, projector.calls)
	}
}

func TestProcessorProjectsNewEvent(t *testing.T) {
	projector := &projectorStub{}
	processed, err := NewProcessor(markerStub{newEvent: true}, projector).Process(context.Background(), events.Envelope{})
	if err != nil || !processed || projector.calls != 1 {
		t.Fatalf("new event should be projected: processed=%v err=%v calls=%d", processed, err, projector.calls)
	}
}

func TestProcessorReturnsProjectionError(t *testing.T) {
	wantErr := errors.New("projection failed")
	projector := &projectorStub{err: wantErr}
	processed, err := NewProcessor(markerStub{newEvent: true}, projector).Process(context.Background(), events.Envelope{})
	if processed || !errors.Is(err, wantErr) {
		t.Fatalf("expected projection error: processed=%v err=%v", processed, err)
	}
}
