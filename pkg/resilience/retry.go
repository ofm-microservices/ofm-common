package resilience

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

// RetryPolicy controls bounded exponential retry behavior.
type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Jitter         time.Duration
}

// DefaultRetryPolicy is safe for transient network and dependency failures.
var DefaultRetryPolicy = RetryPolicy{MaxAttempts: 5, InitialBackoff: 100 * time.Millisecond, MaxBackoff: 30 * time.Second, Jitter: 100 * time.Millisecond}

// Retry executes operation until it succeeds, becomes permanent, the policy
// is exhausted, or the context is cancelled.
func Retry(ctx context.Context, policy RetryPolicy, operation func(context.Context, int) error) error {
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	if policy.InitialBackoff <= 0 {
		policy.InitialBackoff = DefaultRetryPolicy.InitialBackoff
	}
	if policy.MaxBackoff <= 0 {
		policy.MaxBackoff = DefaultRetryPolicy.MaxBackoff
	}
	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		err := operation(ctx, attempt)
		var permanent PermanentError
		if err == nil || errors.As(err, &permanent) {
			return err
		}
		if attempt == policy.MaxAttempts {
			return err
		}
		delay := policy.InitialBackoff * time.Duration(1<<(attempt-1))
		if delay > policy.MaxBackoff {
			delay = policy.MaxBackoff
		}
		if policy.Jitter > 0 {
			delay += time.Duration(rand.Int64N(int64(policy.Jitter)))
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}
