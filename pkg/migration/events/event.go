// Package events defines canonical integration events used by migration
// bridges and legacy projections.
package events

import (
	"encoding/json"
	"time"
)

// Envelope is the versioned metadata shared by every migration event.
type Envelope struct {
	EventID             string          `json:"event_id"`
	CommandID           string          `json:"command_id,omitempty"`
	CorrelationID       string          `json:"correlation_id,omitempty"`
	CausationID         string          `json:"causation_id,omitempty"`
	CommandMethod       string          `json:"command_method,omitempty"`
	CommandPath         string          `json:"command_path,omitempty"`
	IdempotencyKey      string          `json:"idempotency_key,omitempty"`
	RecoveryPrincipalID string          `json:"recovery_principal_id,omitempty"`
	RecoveryUsername    string          `json:"recovery_username,omitempty"`
	RecoveryEmail       string          `json:"recovery_email,omitempty"`
	EventType           string          `json:"event_type"`
	Operation           string          `json:"operation,omitempty"`
	SchemaVersion       int             `json:"schema_version"`
	AggregateType       string          `json:"aggregate_type"`
	AggregateID         string          `json:"aggregate_id"`
	AggregateVersion    int64           `json:"aggregate_version"`
	SourceService       string          `json:"source_service"`
	OccurredAt          time.Time       `json:"occurred_at"`
	Payload             json.RawMessage `json:"payload"`
}

// FallbackMetadata carries request-scoped recovery identity state across the
// Fiber boundary without coupling transport code to persistence packages.
type FallbackMetadata struct {
	CommandID      string
	CorrelationID  string
	IdempotencyKey string
	LegacyIDs      []int64
	ReservedIDs    []int64
}

// SetFallbackMetadata attaches fallback state to contexts such as fasthttp's
// request context, which supports user values while implementing context.Context.
func SetFallbackMetadata(ctx interface{ SetUserValue(any, any) }, metadata *FallbackMetadata) {
	ctx.SetUserValue("ofm.fallback.metadata", metadata)
}

// FallbackMetadataFromContext returns request-scoped fallback state when one
// was attached by the HTTP boundary.
func FallbackMetadataFromContext(ctx interface{ Value(any) any }) (*FallbackMetadata, bool) {
	metadata, ok := ctx.Value("ofm.fallback.metadata").(*FallbackMetadata)
	return metadata, ok && metadata != nil
}

const (
	RegistrationStarted        = "registration.started"
	RegistrationCodeSent       = "registration.code.sent"
	RegistrationEmailVerified  = "registration.email.verified"
	RegistrationCompleted      = "registration.completed"
	RegistrationFailed         = "registration.failed"
	AuthCredentialsCreated     = "auth.credentials.created"
	AuthCredentialsVerified    = "auth.credentials.verified"
	AuthCredentialsDeactivated = "auth.credentials.deactivated"
	AuthRoleAssigned           = "auth.role.assigned"
	UserProfileCreated         = "user.profile.created"
	UserProfileUpdated         = "user.profile.updated"
	UserProfileActivated       = "user.profile.activated"
	UserProfileDeactivated     = "user.profile.deactivated"
)
