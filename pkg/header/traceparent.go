package header

import (
	"strings"
)

// TraceparentHeaderKey is the W3C Trace Context propagation header.
const TraceparentHeaderKey = "traceparent"

// TraceInfo holds resolved trace and span identifiers.
type TraceInfo struct {
	TraceID string // 32 hex, aligned with W3C/OTel
	SpanID  string // 16 hex, may be empty
}

// ParseTraceparent parses a W3C traceparent header value (version 00 only).
func ParseTraceparent(s string) (TraceInfo, bool) {
	parts := strings.Split(s, "-")
	if len(parts) != 4 {
		return TraceInfo{}, false
	}
	if parts[0] != "00" {
		return TraceInfo{}, false
	}
	traceID := parts[1]
	spanID := parts[2]
	flags := parts[3]
	if len(traceID) != 32 || len(spanID) != 16 || len(flags) != 2 {
		return TraceInfo{}, false
	}
	if !isHex(traceID) || !isHex(spanID) || !isHex(flags) {
		return TraceInfo{}, false
	}
	if isAllZero(traceID) {
		return TraceInfo{}, false
	}
	return TraceInfo{
		TraceID: strings.ToLower(traceID),
		SpanID:  strings.ToLower(spanID),
	}, true
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func isAllZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}
