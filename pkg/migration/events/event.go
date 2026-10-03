// Package events defines canonical integration events used by migration
// bridges and legacy projections.
package events

import commonevents "github.com/ofm-microservices/ofm-common/pkg/events"

// Envelope is retained as a compatibility alias for migration code. The
// canonical implementation lives in the shared events package.
type Envelope = commonevents.EventEnvelope

// FallbackMetadata carries request-scoped recovery identity state across the
// Fiber boundary without coupling transport code to persistence packages.
type FallbackMetadata struct {
	CommandID      string
	CorrelationID  string
	IdempotencyKey string
	LegacyIDs      []int64
	ReservedIDs    []int64
	// ReservedUUIDs keeps the external identity chosen by the fallback
	// boundary paired with the reserved legacy identity for each entity.
	ReservedUUIDs map[string]string
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
