package cache

import (
	"time"
)

// Jitter returns ttl with +/- pct jitter (e.g. pct=0.2 means +/-20%).
// It uses a lightweight time-based pseudo randomness to avoid global RNG contention.
func Jitter(ttl time.Duration, pct float64) time.Duration {
	if ttl <= 0 || pct <= 0 {
		return ttl
	}
	delta := int64(float64(ttl) * pct)
	if delta <= 0 {
		return ttl
	}
	// pseudo-random in [-delta, +delta]
	n := time.Now().UnixNano()
	off := (n % (2*delta + 1)) - delta
	return ttl + time.Duration(off)
}

func ClampMin(ttl, min time.Duration) time.Duration {
	if ttl < min {
		return min
	}
	return ttl
}
