package cache

import (
	"context"
	"time"
)

func waitForCacheFill[T any](ctx context.Context, maxWait time.Duration, initialBackoff time.Duration, maxBackoff time.Duration, getter func(context.Context) (T, bool, error)) (T, bool) {
	var zero T
	if maxWait <= 0 {
		return zero, false
	}
	if initialBackoff <= 0 {
		initialBackoff = 20 * time.Millisecond
	}
	if maxBackoff < initialBackoff {
		maxBackoff = initialBackoff
	}

	deadline := time.Now().Add(maxWait)
	wait := initialBackoff
	for time.Now().Before(deadline) {
		value, ok, err := getter(ctx)
		if err == nil && ok {
			return value, true
		}

		select {
		case <-ctx.Done():
			return zero, false
		case <-time.After(wait):
		}

		if wait < maxBackoff {
			wait *= 2
			if wait > maxBackoff {
				wait = maxBackoff
			}
		}
	}

	return zero, false
}
