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
	// BackoffSchedule contains the delay before each subsequent attempt.
	BackoffSchedule []time.Duration
}

// DefaultRetryPolicy is safe for transient network and dependency failures.
var DefaultRetryPolicy = RetryPolicy{
	MaxAttempts:     9,
	InitialBackoff:  time.Second,
	MaxBackoff:      5 * time.Minute,
	BackoffSchedule: []time.Duration{time.Second, 5 * time.Second, 10 * time.Second, 30 * time.Second, time.Minute, 2 * time.Minute, 3 * time.Minute, 5 * time.Minute},
}

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
		delay := BackoffDelay(policy, attempt)
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

// BackoffDelay returns the delay before the retry following attempt.
func BackoffDelay(policy RetryPolicy, attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if len(policy.BackoffSchedule) >= attempt {
		return policy.BackoffSchedule[attempt-1]
	}
	if policy.InitialBackoff <= 0 {
		policy.InitialBackoff = DefaultRetryPolicy.InitialBackoff
	}
	if policy.MaxBackoff <= 0 {
		policy.MaxBackoff = DefaultRetryPolicy.MaxBackoff
	}
	delay := policy.InitialBackoff * time.Duration(1<<(attempt-1))
	if delay > policy.MaxBackoff {
		return policy.MaxBackoff
	}
	return delay
}
