package header

import (
	"net/http"
	"strconv"

	"github.com/ml444/gkit/log"
)

func GetHeader4HTTP(h http.Header, key string) string {
	return h.Get(key)
}

func GetCorpId4Headers(h http.Header) uint64 {
	s := h.Get(CorpIdKey)
	if s == "" {
		return 0
	}
	corpId, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Errorf("err: %v", err)
		return 0
	}
	return corpId
}

func GetUserId4Headers(h http.Header) uint64 {
	s := h.Get(UserIdKey)
	if s == "" {
		return 0
	}
	corpId, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Errorf("err: %v", err)
		return 0
	}
	return corpId
}

func GetTraceId4Headers(h http.Header) (traceId string) {
	return h.Get(TraceIdKey)
}

func SetTraceId2Headers(h http.Header, traceId string) {
	h.Set(TraceIdKey, traceId)
}
