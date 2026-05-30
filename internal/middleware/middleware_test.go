package middleware

import (
	"testing"
	"time"
)

func TestIsLoopbackIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ip: "127.0.0.1", want: true},
		{ip: "127.0.0.1:3456", want: true},
		{ip: "[::1]:8080", want: true},
		{ip: "10.0.0.1", want: false},
		{ip: "192.168.1.10:1234", want: false},
	}

	for _, tt := range tests {
		if got := IsLoopbackIP(tt.ip); got != tt.want {
			t.Errorf("IsLoopbackIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestRateLimiter_LoopbackHasHigherBudget(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	for i := 0; i < 10; i++ {
		if !rl.Allow("127.0.0.1:54321") {
			t.Fatalf("loopback request %d should be allowed", i+1)
		}
	}

	// 11th request within the same window should still be allowed for loopback.
	if !rl.Allow("127.0.0.1:54321") {
		t.Fatal("loopback should use the higher loopback budget")
	}
}

func TestRateLimiter_NonLoopbackUsesDefaultBudget(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	if !rl.Allow("203.0.113.10:1234") {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("203.0.113.10:1234") {
		t.Fatal("second request should be allowed")
	}
	if rl.Allow("203.0.113.10:1234") {
		t.Fatal("third request should be rate limited")
	}
}

func TestRequestDeduplicator_ReleaseAllowsRetry(t *testing.T) {
	d := NewRequestDeduplicator(time.Second)
	body := []byte(`{"model":"test"}`)

	if _, ok := d.TryAcquire(body); !ok {
		t.Fatal("first acquire should succeed")
	}
	if _, ok := d.TryAcquire(body); ok {
		t.Fatal("duplicate should be rejected while in flight")
	}

	d.Release(body)

	if _, ok := d.TryAcquire(body); !ok {
		t.Fatal("acquire should succeed after release")
	}
}
