package ratelimit

import (
	"strings"
	"time"
)

// DefaultKeyPrefix is the Redis key namespace for distributed rate limiting.
const DefaultKeyPrefix = "gkit:rl"

// RateLimitKey builds a store key: gkit:rl:{service}:{path}:{windowMs}.
// Path segments are normalized (leading slash, no spaces).
func RateLimitKey(service, path string, period time.Duration) string {
	service = sanitizeKeyPart(service)
	if service == "" {
		service = "default"
	}
	path = sanitizeKeyPart(path)
	if path == "" {
		path = "_"
	}
	window := period.Milliseconds()
	if window <= 0 {
		window = 1
	}
	return DefaultKeyPrefix + ":" + service + ":" + path + ":" + itoa(window)
}

func sanitizeKeyPart(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.TrimPrefix(s, "/")
	return s
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
