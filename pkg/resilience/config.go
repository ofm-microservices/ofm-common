package resilience

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// RetryPolicyFromEnv loads the shared retry policy from process configuration.
func RetryPolicyFromEnv() RetryPolicy {
	policy := DefaultRetryPolicy
	policy.MaxAttempts = envInt("RETRY_MAX_ATTEMPTS", policy.MaxAttempts)
	policy.InitialBackoff = envDuration("RETRY_INITIAL_BACKOFF", policy.InitialBackoff)
	policy.MaxBackoff = envDuration("RETRY_MAX_BACKOFF", policy.MaxBackoff)
	if raw := strings.TrimSpace(os.Getenv("RETRY_BACKOFF_SCHEDULE")); raw != "" {
		policy.BackoffSchedule = parseSchedule(raw)
		if len(policy.BackoffSchedule) > 0 && os.Getenv("RETRY_MAX_ATTEMPTS") == "" {
			policy.MaxAttempts = len(policy.BackoffSchedule) + 1
		}
	}
	return policy
}

func parseSchedule(raw string) []time.Duration {
	parts := strings.Split(raw, ",")
	schedule := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		if duration, err := time.ParseDuration(strings.TrimSpace(part)); err == nil && duration > 0 {
			schedule = append(schedule, duration)
		}
	}
	return schedule
}

// BreakerConfigFromEnv loads circuit-breaker settings from process configuration.
func BreakerConfigFromEnv() BreakerConfig {
	return BreakerConfig{
		FailureThreshold: envInt("CIRCUIT_BREAKER_FAILURE_THRESHOLD", 5),
		ResetTimeout:     envDuration("CIRCUIT_BREAKER_RESET_TIMEOUT", 30*time.Second),
		CallTimeout:      envDuration("CIRCUIT_BREAKER_CALL_TIMEOUT", 5*time.Second),
	}
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
