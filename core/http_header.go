package core

import (
	"github.com/ml444/gkit/log"
	"net/http"
	"strconv"
)

const (
	HttpHeaderUserId     = "x-user-id"
	HttpHeaderCorpId     = "x-corp-id"
	HttpHeaderClientType = "x-client-type"
	HttpHeaderClientId   = "x-client-id"
)

func GetHeader4HTTP(h http.Header, key string) string {
	return h.Get(key)
}

func GetCorpId4Headers(h http.Header) uint64 {
	s := h.Get(HttpHeaderCorpId)
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
	s := h.Get(HttpHeaderUserId)
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
