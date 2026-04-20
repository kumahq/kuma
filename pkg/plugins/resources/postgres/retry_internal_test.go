package postgres

import (
	"context"
	"errors"
	"testing"
)

type safeToRetryErr struct{ safe bool }

func (e *safeToRetryErr) Error() string     { return "conn closed" }
func (e *safeToRetryErr) SafeToRetry() bool { return e.safe }

type wrappedErr struct{ inner error }

func (w *wrappedErr) Error() string { return "wrapped: " + w.inner.Error() }
func (w *wrappedErr) Unwrap() error { return w.inner }

func TestRetryOnSafeToRetry(t *testing.T) {
	t.Parallel()

	t.Run("returns nil on success", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			return nil
		})
		if err != nil {
			t.Fatalf("want nil, got %v", err)
		}
		if calls != 1 {
			t.Fatalf("want 1 call, got %d", calls)
		}
	})

	t.Run("retries safe-to-retry errors and eventually succeeds", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			if calls < 3 {
				return &safeToRetryErr{safe: true}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("want nil, got %v", err)
		}
		if calls != 3 {
			t.Fatalf("want 3 calls, got %d", calls)
		}
	})

	t.Run("does not retry non-safe errors", func(t *testing.T) {
		calls := 0
		sentinel := errors.New("not retryable")
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("want sentinel, got %v", err)
		}
		if calls != 1 {
			t.Fatalf("want 1 call, got %d", calls)
		}
	})

	t.Run("does not retry errors where SafeToRetry returns false", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			return &safeToRetryErr{safe: false}
		})
		if err == nil {
			t.Fatalf("want error, got nil")
		}
		if calls != 1 {
			t.Fatalf("want 1 call, got %d", calls)
		}
	})

	t.Run("unwraps errors via errors.As", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			if calls < 2 {
				return &wrappedErr{inner: &safeToRetryErr{safe: true}}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("want nil, got %v", err)
		}
		if calls != 2 {
			t.Fatalf("want 2 calls, got %d", calls)
		}
	})

	t.Run("skips retry when in transaction", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), true, func() error {
			calls++
			return &safeToRetryErr{safe: true}
		})
		if err == nil {
			t.Fatalf("want error, got nil")
		}
		if calls != 1 {
			t.Fatalf("want 1 call (no retry inside tx), got %d", calls)
		}
	})

	t.Run("gives up after max retries", func(t *testing.T) {
		calls := 0
		err := retryOnSafeToRetry(context.Background(), false, func() error {
			calls++
			return &safeToRetryErr{safe: true}
		})
		if err == nil {
			t.Fatalf("want error, got nil")
		}
		// 1 initial + 3 retries = 4 calls
		if calls != 4 {
			t.Fatalf("want 4 calls, got %d", calls)
		}
	})
}
