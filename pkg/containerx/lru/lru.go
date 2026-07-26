//go:build go1.9

package lru

import (
	"container/list"
	"sync"
)

type LRUCache[K comparable, V any] struct {
	size      int
	innerList *list.List
	innerMap  sync.Map
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](size int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		size:      size,
		innerList: list.New(),
		innerMap:  sync.Map{},
	}
}

func (lru *LRUCache[K, V]) loadElement(key K) (*list.Element, bool) {
	v, ok := lru.innerMap.Load(key)
	if !ok {
		return nil, false
	}
	el, ok := v.(*list.Element)
	if !ok {
		return nil, false
	}
	return el, true
}

func (lru *LRUCache[K, V]) storeElement(key K, el *list.Element) {
	lru.innerMap.Store(key, el)
}

func (lru *LRUCache[K, V]) deleteElement(key K) {
	lru.innerMap.Delete(key)
}

func (lru *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if e, ok := lru.loadElement(key); ok {
		lru.innerList.MoveToFront(e)
		return e.Value.(*entry[K, V]).value, true
	}
	return
}

func (lru *LRUCache[K, V]) Put(key K, value V) (evicted bool) {
	if el, ok := lru.loadElement(key); ok {
		lru.innerList.MoveToFront(el)
		el.Value.(*entry[K, V]).value = value
		return false
	} else {
		e := &entry[K, V]{key, value}
		el = lru.innerList.PushFront(e)
		lru.storeElement(key, el)

		if lru.innerList.Len() > lru.size {
			last := lru.innerList.Back()
			lru.innerList.Remove(last)
			lru.deleteElement(last.Value.(*entry[K, V]).key)
			return true
		}
		return false
	}
}
