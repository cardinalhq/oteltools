package chqpb

import (
	"container/heap"
	"sort"
	"sync"
	"time"
)

type itemWithValue struct {
	Tid      int64
	Value    float64
	LastSeen time.Time
}

type valueMinHeap struct {
	items []itemWithValue
	index map[int64]int
}

func (h valueMinHeap) Len() int           { return len(h.items) }
func (h valueMinHeap) Less(i, j int) bool { return h.items[i].Value < h.items[j].Value }
func (h valueMinHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i].Tid] = i
	h.index[h.items[j].Tid] = j
}

func (h *valueMinHeap) Push(x interface{}) {
	item := x.(itemWithValue)
	h.index[item.Tid] = len(h.items)
	h.items = append(h.items, item)
}

func (h *valueMinHeap) Pop() interface{} {
	n := len(h.items)
	item := h.items[n-1]
	h.items = h.items[0 : n-1]
	delete(h.index, item.Tid)
	return item
}

type TopKByValue struct {
	mu    sync.RWMutex
	k     int
	items map[int64]float64
	h     *valueMinHeap
	ttl   time.Duration
}

func NewTopKByValue(k int, ttl time.Duration) *TopKByValue {
	return &TopKByValue{
		k:     k,
		items: make(map[int64]float64, k),
		h: &valueMinHeap{
			items: make([]itemWithValue, 0, k),
			index: make(map[int64]int),
		},
		ttl: ttl,
	}
}

func (t *TopKByValue) Add(tid int64, value float64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	// update existing
	if existing, ok := t.items[tid]; ok {
		if value <= existing {
			return ok
		}
		t.items[tid] = value
		idx := t.h.index[tid]
		t.h.items[idx].Value = value
		t.h.items[idx].LastSeen = now
		heap.Fix(t.h, idx)
		return true
	}
	// add if capacity
	if len(t.h.items) < t.k {
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}
	// replace min if better
	if value > t.h.items[0].Value {
		evicted := heap.Pop(t.h).(itemWithValue)
		delete(t.items, evicted.Tid)
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}
	return false
}

func (t *TopKByValue) Eligible(value float64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.h.items) < t.k {
		return true
	}
	return len(t.h.items) > 0 && value > t.h.items[0].Value
}

func (t *TopKByValue) CleanupExpired() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	j := 0
	for i := 0; i < len(t.h.items); i++ {
		item := t.h.items[i]
		if now.Sub(item.LastSeen) <= t.ttl {
			t.h.items[j] = item
			t.h.index[item.Tid] = j
			j++
		} else {
			delete(t.items, item.Tid)
		}
	}
	t.h.items = t.h.items[:j]
	heap.Init(t.h)
}

func (t *TopKByValue) SortedSlice() []itemWithValue {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]itemWithValue, len(t.h.items))
	copy(out, t.h.items)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Value > out[j].Value
	})
	return out
}
