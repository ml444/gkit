package minheap

import (
	"testing"
	"time"
)

func TestNewMinHeap(t *testing.T) {
	mh := NewMinHeap[*Element[int]](10)
	for i := 0; i < 15; i++ {
		err := mh.PushEl(&Element[int]{
			Value:    123 + i,
			priority: time.Now().Unix(),
			index:    i,
		})
		if err != nil {
			t.Log(err.Error())
		}
		t.Log("===>", i)
		time.Sleep(time.Second * 1)
	}
	for i := 0; i < 15; i++ {
		el := mh.PopEl()
		if el == nil {
			continue
		}

		t.Log("===> ", el.index, el.Priority())
	}
}
