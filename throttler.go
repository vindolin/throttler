package throttler

import (
	"sync"
	"time"
)

// throttler is a concurrency-safe throttler that can be used to limit the rate
// at which certain operations are performed. It uses a map to store the last
// time each unique value was allowed.
type throttler struct {
	// mutex protects the valueMap
	mu sync.Mutex
	// valueMap maps each unique value to the last time it was allowed
	valueMap map[interface{}]time.Time
	// throttle is the minimum duration between allows for each value
	throttle time.Duration
	// cleanup is the interval at which expired entries are removed from the map
	cleanup time.Duration
	// lastClean is the last time the map was cleaned
	lastClean time.Time
}

// NewThrottler creates a new NewThrottler that will allow at most one request every
// throttle duration, and will expire entries after cleanup has passed.
func NewThrottler(throttle, cleanup time.Duration) *throttler {
	return &throttler{
		valueMap:  make(map[interface{}]time.Time),
		throttle:  throttle,
		cleanup:   cleanup,
		lastClean: time.Now(),
	}
}

// Allow checks if the given value is allowed to do an operation. If the value
// is not present in the map or if the last time it was allowed is more than
// throttle duration ago, the value is added to the map and true is returned.
// Otherwise, false is returned.
func (t *throttler) Allow(value interface{}) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clean up old entries if the cleanup interval has passed
	if time.Since(t.lastClean) > t.cleanup {
		for val, lastTime := range t.valueMap {
			if time.Since(lastTime) > t.throttle {
				delete(t.valueMap, val)
			}
		}
		t.lastClean = time.Now()
	}

	// Check if request is allowed
	if lastTime, ok := t.valueMap[value]; ok && time.Since(lastTime) < t.throttle {
		return false
	}

	t.valueMap[value] = time.Now()
	return true
}
