package resilience

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

// BreakerConfig controls circuit state transitions.
type BreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	CallTimeout      time.Duration
}

// State describes the current circuit state for telemetry and operations.
type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half_open"
)

// Breaker is a concurrency-safe circuit breaker for remote dependencies.
type Breaker struct {
	mu       sync.Mutex
	config   BreakerConfig
	failures int
	openedAt time.Time
}

// NewBreaker creates a closed circuit with production-safe defaults.
func NewBreaker(config BreakerConfig) *Breaker {
	if config.FailureThreshold < 1 {
		config.FailureThreshold = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 30 * time.Second
	}
	if config.CallTimeout <= 0 {
		config.CallTimeout = 5 * time.Second
	}
	return &Breaker{config: config}
}

// Do executes operation unless the circuit is open. Successful calls close
// the circuit and failed calls contribute to opening it.
func (b *Breaker) Do(ctx context.Context, operation func(context.Context) error) error {
	if !b.allow() {
		return ErrCircuitOpen
	}
	callCtx, cancel := context.WithTimeout(ctx, b.config.CallTimeout)
	err := operation(callCtx)
	cancel()
	b.record(err)
	return err
}

// DoClassified executes an operation and trips the circuit only when classify
// identifies a dependency failure. Callers that wrap databases or brokers can
// therefore keep validation, not-found, and constraint errors local to the
// request instead of poisoning the circuit for unrelated requests.
func (b *Breaker) DoClassified(ctx context.Context, operation func(context.Context) error, classify func(error) bool) error {
	if !b.allow() {
		return ErrCircuitOpen
	}
	callCtx, cancel := context.WithTimeout(ctx, b.config.CallTimeout)
	err := operation(callCtx)
	cancel()
	if err == nil || classify == nil || classify(err) {
		b.record(err)
	}
	return err
}

// IsTransientDependencyError identifies failures for which a circuit breaker
// should stop sending traffic: context deadlines/cancellation, network
// failures, and PostgreSQL-style SQLSTATE connection exceptions (08xxx).
// Application SQL errors remain visible to the caller but do not open the
// dependency circuit.
func IsTransientDependencyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var stateErr interface{ SQLState() string }
	if errors.As(err, &stateErr) {
		return strings.HasPrefix(stateErr.SQLState(), "08")
	}
	return false
}

// State returns the current operational state of the circuit.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return StateClosed
	}
	if time.Since(b.openedAt) >= b.config.ResetTimeout {
		return StateHalfOpen
	}
	return StateOpen
}

func (b *Breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return true
	}
	if time.Since(b.openedAt) < b.config.ResetTimeout {
		return false
	}
	b.openedAt = time.Time{}
	b.failures = 0
	return true
}

func (b *Breaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err == nil {
		b.failures = 0
		b.openedAt = time.Time{}
		return
	}
	b.failures++
	if b.failures >= b.config.FailureThreshold {
		b.openedAt = time.Now()
	}
}
