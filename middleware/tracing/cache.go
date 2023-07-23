package tracing

import (
	"fmt"
	"sync"

	"github.com/petermattis/goid"
)

var CacheTraceId *traceIdCache

func init() {
	CacheTraceId = NewCacheTraceId()
}

type traceIdCache struct {
	TraceIdMap map[int64]string
	mux        sync.RWMutex
}

func NewCacheTraceId() *traceIdCache {
	return &traceIdCache{
		TraceIdMap: map[int64]string{},
		mux:        sync.RWMutex{},
	}
}

func (c *traceIdCache) GetTraceId(key int64) string {
	if key == 0 {
		key = goid.Get()
	}
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.TraceIdMap[key]
}

func (c *traceIdCache) SetTraceId(key int64, value string) {
	if key == 0 {
		key = goid.Get()
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if value == "" {
		// delete
		delete(c.TraceIdMap, key)
	} else {
		c.TraceIdMap[key] = value
	}
}

func (c *traceIdCache) GetTraceIdWithoutKey() string {
	return c.GetTraceId(0)
}
func (c *traceIdCache) SetTraceIdWithoutKey(value string) {
	c.SetTraceId(0, value)
}

func GetTraceIdFromCache(key int64) string {
	if key == 0 {
		key = goid.Get()
	}
	traceId := CacheTraceId.GetTraceId(key)
	if traceId == "" {
		fmt.Println("[", key, "] traceId is empty")
	}
	return traceId
}
