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

	// Check if item already exists
	if existing, ok := t.items[tid]; ok {
		if value <= existing {
			return true // Already have better or equal value
		}

		t.items[tid] = value

		// Validate index before accessing heap items
		if idx, indexOk := t.h.index[tid]; indexOk {
			if idx < 0 || idx >= len(t.h.items) {
				// Index is corrupted - rebuild the index
				t.rebuildIndex()
				// Try to find the item again after rebuilding
				if newIdx, stillOk := t.h.index[tid]; stillOk && newIdx >= 0 && newIdx < len(t.h.items) {
					t.h.items[newIdx].Value = value
					t.h.items[newIdx].LastSeen = now
					heap.Fix(t.h, newIdx)
					return true
				}
				// If still not found after rebuild, treat as new item (fallthrough)
			} else {
				// Valid index - update the item
				t.h.items[idx].Value = value
				t.h.items[idx].LastSeen = now
				heap.Fix(t.h, idx)
				return true
			}
		}

		// If we reach here, the item exists in the items map but not in heap
		// This shouldn't happen in normal operation - rebuild and continue
		t.rebuildIndex()
	}

	// Add if we have capacity
	if len(t.h.items) < t.k {
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}

	// Replace minimum if this value is better
	if len(t.h.items) > 0 && value > t.h.items[0].Value {
		evicted := heap.Pop(t.h).(itemWithValue)
		delete(t.items, evicted.Tid)
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}

	return false
}

// rebuildIndex reconstructs the index map from the current heap items
func (t *TopKByValue) rebuildIndex() {
	t.h.index = make(map[int64]int, len(t.h.items))
	for i, item := range t.h.items {
		t.h.index[item.Tid] = i
	}
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

	// Only reinitialize heap if we have items
	if len(t.h.items) > 0 {
		heap.Init(t.h)
	} else {
		// Clear the index if no items remain
		t.h.index = make(map[int64]int)
	}
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
