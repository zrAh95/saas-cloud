package services

import (
	"sync"
	"time"
)

type RateLimitEntry struct {
	Count     int
	LastReset time.Time
}

var rateLimitMu sync.Mutex
var rateLimitMap = make(map[string]*RateLimitEntry)

// CheckRateLimit checks if the action is allowed based on rate limiting
// Returns true if action is allowed, false if rate limit exceeded
// key: identifier (email, IP, etc)
// maxAttempts: maximum attempts allowed
// windowSeconds: time window in seconds
func CheckRateLimit(key string, maxAttempts int, windowSeconds int) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	now := time.Now()
	entry, exists := rateLimitMap[key]

	if !exists {
		// First attempt
		rateLimitMap[key] = &RateLimitEntry{
			Count:     1,
			LastReset: now,
		}
		return true
	}

	// Check if time window has expired
	if time.Since(entry.LastReset) > time.Duration(windowSeconds)*time.Second {
		// Reset counter
		rateLimitMap[key] = &RateLimitEntry{
			Count:     1,
			LastReset: now,
		}
		return true
	}

	// Within time window
	if entry.Count < maxAttempts {
		entry.Count++
		return true
	}

	return false
}

// ResetRateLimit resets rate limit counter for a key (use after successful login)
func ResetRateLimit(key string) {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()
	delete(rateLimitMap, key)
}

// GetRateLimitRemaining returns how many attempts remaining
// windowSeconds: time window in seconds (must match the window used in CheckRateLimit)
func GetRateLimitRemaining(key string, maxAttempts int, windowSeconds int) int {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	entry, exists := rateLimitMap[key]
	if !exists {
		return maxAttempts
	}

	if time.Since(entry.LastReset) > time.Duration(windowSeconds)*time.Second {
		return maxAttempts
	}

	remaining := maxAttempts - entry.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}
