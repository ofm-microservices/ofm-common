package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryStopsAtSuccess(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), RetryPolicy{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: time.Millisecond}, func(context.Context, int) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil || attempts != 2 {
		t.Fatalf("retry attempts=%d err=%v", attempts, err)
	}
}

func TestPermanentErrorSkipsRetry(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), DefaultRetryPolicy, func(context.Context, int) error {
		attempts++
		return Permanent(errors.New("invalid payload"))
	})
	if err == nil || attempts != 1 {
		t.Fatalf("expected one permanent attempt, attempts=%d err=%v", attempts, err)
	}
}

func TestBreakerOpensAndResets(t *testing.T) {
	b := NewBreaker(BreakerConfig{FailureThreshold: 2, ResetTimeout: time.Millisecond, CallTimeout: time.Second})
	fail := func(context.Context) error { return errors.New("down") }
	_ = b.Do(context.Background(), fail)
	_ = b.Do(context.Background(), fail)
	if !errors.Is(b.Do(context.Background(), fail), ErrCircuitOpen) {
		t.Fatal("expected open circuit")
	}
	time.Sleep(2 * time.Millisecond)
	if err := b.Do(context.Background(), func(context.Context) error { return nil }); err != nil {
		t.Fatalf("breaker did not recover: %v", err)
	}
}

func TestClassifiedBreakerIgnoresApplicationErrors(t *testing.T) {
	b := NewBreaker(BreakerConfig{FailureThreshold: 2, ResetTimeout: time.Second, CallTimeout: time.Second})
	applicationErr := errors.New("unique constraint violation")
	for range 5 {
		if err := b.DoClassified(context.Background(), func(context.Context) error { return applicationErr }, IsTransientDependencyError); !errors.Is(err, applicationErr) {
			t.Fatalf("classified call returned %v", err)
		}
	}
	if got := b.State(); got != StateClosed {
		t.Fatalf("application errors opened dependency breaker: %s", got)
	}
}
