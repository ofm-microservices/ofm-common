package realtime

import (
	"context"
	"testing"

	"github.com/ofm-microservices/ofm-common/pkg/observability/metadata"
)

func TestNewUsesRequestCorrelationAndIdempotency(t *testing.T) {
	ctx := metadata.WithValues(context.Background(), metadata.Values{
		RequestID: "request-1", CorrelationID: "correlation-1", IdempotencyKey: "operation-1",
	})
	notification := New(ctx, "gig.published", "gig", "gig-1", "user-1", nil)
	if notification.OperationID != "operation-1" {
		t.Fatalf("unexpected operation id: %q", notification.OperationID)
	}
	if notification.CorrelationID != "correlation-1" || notification.Status != StatusAccepted {
		t.Fatalf("unexpected correlation/status: %#v", notification)
	}
	if notification.EventID == "" || notification.OccurredAt == "" {
		t.Fatalf("notification is missing identity/timestamp: %#v", notification)
	}
}
