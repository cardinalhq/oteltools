package chqpb

import (
	"container/heap"
	"sort"
	"sync"
)

// itemWithValue represents a TID and its associated float64 value.
type itemWithValue struct {
	Tid   int64
	Value float64
}

// minHeap implements heap.Interface and holds itemWithValue structs in ascending order of Value.
type minHeap []itemWithValue

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].Value < h[j].Value }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(itemWithValue))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

// TopKByValue maintains a thread-safe top-K collection of TIDs by their float64 values.
type TopKByValue struct {
	mu    sync.RWMutex
	k     int
	items map[int64]float64
	h     minHeap
}

// NewTopKByValue creates a new TopKByValue for up to k items.
func NewTopKByValue(k int) *TopKByValue {
	return &TopKByValue{
		k:     k,
		items: make(map[int64]float64, k),
		h:     make(minHeap, 0, k),
	}
}

// Add considers tid with the given value for top-K ranking.
// If tid is already present with a lower value, it is updated and heap.Fix is called.
// If tid is new and the heap is under capacity, it is pushed.
// If heap is full and value > min, the smallest is evicted and tid is added.
func (t *TopKByValue) Add(tid int64, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update existing entry if value improved
	if existing, ok := t.items[tid]; ok {
		if value <= existing {
			return // no improvement
		}
		// update stored value
		t.items[tid] = value
		// find and fix in heap
		for i := range t.h {
			if t.h[i].Tid == tid {
				t.h[i].Value = value
				heap.Fix(&t.h, i)
				return
			}
		}
		return
	}

	// Insert new entry if under capacity
	if len(t.items) < t.k {
		heap.Push(&t.h, itemWithValue{Tid: tid, Value: value})
		t.items[tid] = value
		return
	}

	// Otherwise, only insert if value beats current minimum
	if len(t.h) > 0 && value > t.h[0].Value {
		// evict smallest
		evicted := heap.Pop(&t.h).(itemWithValue)
		delete(t.items, evicted.Tid)

		// add new
		heap.Push(&t.h, itemWithValue{Tid: tid, Value: value})
		t.items[tid] = value
	}
}

// Eligible reports whether tid should be allowed into the top-K set.
//   - Before we’ve seen k distinct tids, everyone is eligible.
//   - Once full, only those already in the top-K remain eligible.
func (t *TopKByValue) Eligible(value float64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.items) < t.k {
		return true
	}
	return len(t.h) > 0 && value > t.h[0].Value
}

// SortedSlice returns the current top-K items sorted descending by value.
func (t *TopKByValue) SortedSlice() []itemWithValue {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]itemWithValue, len(t.h))
	copy(out, t.h)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Value > out[j].Value
	})
	return out
}
