package ratelimit

import (
	"testing"
	"time"
)

func TestRateLimitKey(t *testing.T) {
	got := RateLimitKey("user-svc", "/api/v1/users/{id}", time.Second)
	want := "gkit:rl:user-svc:api/v1/users/{id}:1000"
	if got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

func TestRateLimitKeyDefaultService(t *testing.T) {
	got := RateLimitKey("", "/health", 500*time.Millisecond)
	if got != "gkit:rl:default:health:500" {
		t.Fatalf("key = %q", got)
	}
}
