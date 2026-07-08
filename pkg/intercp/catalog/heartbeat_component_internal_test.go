package catalog

import (
	"testing"
	"time"
)

func TestHeartbeatBackoffDoesNotReduceConfiguredInterval(t *testing.T) {
	testCases := []struct {
		name     string
		interval time.Duration
		failures int
		expected time.Duration
	}{
		{
			name:     "successful heartbeat uses configured interval",
			interval: 10 * time.Second,
			failures: 0,
			expected: 10 * time.Second,
		},
		{
			name:     "failed heartbeat backs off below cap",
			interval: 10 * time.Second,
			failures: 1,
			expected: 20 * time.Second,
		},
		{
			name:     "failed heartbeat is capped at max backoff",
			interval: 10 * time.Second,
			failures: 4,
			expected: maxHeartbeatBackoff,
		},
		{
			name:     "configured interval above max backoff is preserved",
			interval: 2 * time.Minute,
			failures: 1,
			expected: 2 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &heartbeatComponent{
				interval: tc.interval,
				failures: tc.failures,
			}

			if got := h.backoff(); got != tc.expected {
				t.Fatalf("expected backoff %s, got %s", tc.expected, got)
			}
		})
	}
}
