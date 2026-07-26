package minheap

import (
	"container/heap"
	"errors"
)

// An Element is something we manage in a priority queue.
type Element[V any] struct {
	Value    V
	priority int64 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.

	Count int // The number of times the push is repeated
}

func (e *Element[V]) Priority() int64 {
	return e.priority
}

func (e *Element[V]) SetPriority(p int64) {
	e.priority = p
}
func (e *Element[V]) Index() int {
	return e.index
}
func (e *Element[V]) SetIndex(idx int) {
	e.index = idx
}

type IElement interface {
	Priority() int64
	SetPriority(n int64)
	Index() int
	SetIndex(idx int)
}

// MinHeap A PriorityQueue implements heap.Interface and holds Items.
type MinHeap[E IElement] struct {
	list    []E
	maxSize int
}

func NewMinHeap[E IElement](size int) *MinHeap[E] {
	mh := &MinHeap[E]{maxSize: size}
	heap.Init(mh)
	return mh
}

func (mh *MinHeap[E]) Len() int { return len(mh.list) }

func (mh *MinHeap[E]) Less(i, j int) bool {
	return mh.list[i].Priority() < mh.list[j].Priority()
}

func (mh *MinHeap[E]) Swap(i, j int) {
	mh.list[i], mh.list[j] = mh.list[j], mh.list[i]
	mh.list[i].SetIndex(i)
	mh.list[j].SetIndex(j)
}

func (mh *MinHeap[E]) Push(x interface{}) {
	mh.list = append(mh.list, x.(E))
}

func (mh *MinHeap[E]) Pop() any {
	old := mh.list
	n := len(old)
	item := old[n-1]
	//item.index = -1 // for safety
	mh.list = old[0 : n-1]
	return item
}

func (mh *MinHeap[E]) PushEl(el E) error {
	n := len(mh.list)
	if mh.maxSize > 0 && n >= mh.maxSize {
		return errors.New("heap max size")
	}
	el.SetIndex(n)
	heap.Push(mh, el)
	return nil
}

func (mh *MinHeap[E]) PopEl() E {
	if len(mh.list) == 0 {
		var zero E
		return zero
	}
	//el := heap.Pop(mh)
	//return el.(E)
	old := mh.list
	n := len(old)
	item := old[n-1]
	mh.list = old[0 : n-1]
	return item
}

func (mh *MinHeap[E]) PeekMaxEl() E {
	if len(mh.list) == 0 {
		var zero E
		return zero
	}
	return mh.list[0]
}

func (mh *MinHeap[E]) PeekMinEl() E {
	if len(mh.list) == 0 {
		var zero E
		return zero
	}
	return mh.list[mh.Len()-1]
}

// UpdateEl update modifies the priority and value of an Item in the queue.
func (mh *MinHeap[E]) UpdateEl(el E, priority int64) {
	heap.Remove(mh, el.Index())
	el.SetPriority(priority)
	heap.Push(mh, el)
}

func (mh *MinHeap[E]) RemoveEl(el E) {
	heap.Remove(mh, el.Index())
}
