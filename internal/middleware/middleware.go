// Package middleware provides HTTP middleware for the proxy.
package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultRateLimit is the per-IP request budget for non-loopback clients.
	DefaultRateLimit = 500
	// LoopbackRateLimit is higher because Claude Code agents burst many parallel
	// requests to the local proxy on 127.0.0.1.
	LoopbackRateLimit = 2000
)

// RequestDeduplicator prevents duplicate requests from flooding the upstream.
// Uses a hash of the request body to detect duplicates within a time window.
type RequestDeduplicator struct {
	inFlight    sync.Map // request hash → context cancel func
	dedupWindow time.Duration
	logger      *slog.Logger
}

// NewRequestDeduplicator creates a new request deduplicator.
func NewRequestDeduplicator(window time.Duration) *RequestDeduplicator {
	if window == 0 {
		window = 500 * time.Millisecond
	}
	return &RequestDeduplicator{
		dedupWindow: window,
		logger:      slog.Default(),
	}
}

// hashRequest creates a hash from the request body for deduplication.
func hashRequest(body json.RawMessage) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// TryAcquire attempts to acquire a deduplication slot for a request.
// Returns a context to use for the request, and true if this is a new request.
// If false, the returned context is already cancelled and the caller should wait.
func (d *RequestDeduplicator) TryAcquire(body json.RawMessage) (context.Context, bool) {
	hash := hashRequest(body)

	// Check if request is already in flight
	if _, exists := d.inFlight.Load(hash); exists {
		d.logger.Debug("duplicate request detected, waiting", "hash", hash[:8])
		return nil, false
	}

	// Create a cancellable context for this request
	ctx, cancel := context.WithCancel(context.Background())

	// Store the cancel function
	d.inFlight.Store(hash, cancel)

	// Auto-release after window
	go func() {
		select {
		case <-time.After(d.dedupWindow):
			d.Release(body)
		case <-ctx.Done():
			// Context was cancelled, release immediately
		}
	}()

	return ctx, true
}

// Release releases a deduplication slot.
func (d *RequestDeduplicator) Release(body json.RawMessage) {
	hash := hashRequest(body)
	if cancel, exists := d.inFlight.LoadAndDelete(hash); exists {
		cancel.(context.CancelFunc)()
	}
}

// RateLimiter provides per-client IP rate limiting.
type RateLimiter struct {
	tokens map[string]*clientTokenBucket
	mu     sync.RWMutex
	rate   int // tokens per window
	window time.Duration
	logger *slog.Logger
}

// clientTokenBucket holds rate limit state for a single client.
type clientTokenBucket struct {
	tokens   int
	lastFill time.Time
}

// IsLoopbackIP reports whether the client address is a loopback interface.
func IsLoopbackIP(clientIP string) bool {
	host := clientIP
	if h, _, err := net.SplitHostPort(clientIP); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// rateLimitForClient returns the request budget for the given client address.
func rateLimitForClient(clientIP string, defaultRate int) int {
	if IsLoopbackIP(clientIP) {
		return LoopbackRateLimit
	}
	if defaultRate <= 0 {
		return DefaultRateLimit
	}
	return defaultRate
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	if rate <= 0 {
		rate = DefaultRateLimit
	}
	if window == 0 {
		window = time.Minute
	}
	return &RateLimiter{
		tokens: make(map[string]*clientTokenBucket),
		rate:   rate,
		window: window,
		logger: slog.Default(),
	}
}

// Allow checks if a request from the given IP is allowed.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) Allow(clientIP string) bool {
	limit := rateLimitForClient(clientIP, rl.rate)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.tokens[clientIP]

	if !exists {
		rl.tokens[clientIP] = &clientTokenBucket{
			tokens:   limit - 1,
			lastFill: now,
		}
		return true
	}

	// Refill tokens if window has passed
	elapsed := now.Sub(bucket.lastFill)
	if elapsed >= rl.window {
		bucket.tokens = limit
		bucket.lastFill = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	rl.logger.Warn("rate limited", "client", clientIP, "remaining", bucket.tokens)
	return false
}

// GetClientIP extracts the client IP from an HTTP request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (if behind a proxy)
	// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
	// We want the first (leftmost) IP which is the original client.
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RequestIDGenerator generates unique request IDs.
type RequestIDGenerator struct {
	mu      sync.Mutex
	counter uint64
}

// NewRequestIDGenerator creates a new request ID generator.
func NewRequestIDGenerator() *RequestIDGenerator {
	return &RequestIDGenerator{}
}

// Generate creates a new unique request ID.
func (g *RequestIDGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter++
	return fmt.Sprintf("req-%d-%d", time.Now().Unix(), g.counter)
}
