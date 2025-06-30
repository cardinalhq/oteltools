// Copyright 2024-2025 CardinalHQ, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chqpb

import (
	"container/heap"
	"sort"
	"sync"
	"time"
)

// Direction indicates whether we want the top-K highest values (UP)
// or the top-K lowest values (DOWN).
// In proto3, you should have defined:
//   enum Direction { DIRECTION_UNSPECIFIED = 0; UP = 1; DOWN = 2; }
// and generated the Go type accordingly.

type itemWithValue struct {
	Tid      int64
	Value    float64
	LastSeen time.Time
}

// valueHeap is a heap that orders items according to dir:
// for UP: min-heap (smallest at root), evict smallest to keep top K largest;
// for DOWN: max-heap (largest at root), evict largest to keep bottom K smallest.
type valueHeap struct {
	items []itemWithValue
	index map[int64]int
	dir   Direction
}

func (h valueHeap) Len() int { return len(h.items) }
func (h valueHeap) Less(i, j int) bool {
	if h.dir == Direction_UP {
		return h.items[i].Value < h.items[j].Value
	}
	// DOWN
	return h.items[i].Value > h.items[j].Value
}
func (h valueHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i].Tid] = i
	h.index[h.items[j].Tid] = j
}

func (h *valueHeap) Push(x interface{}) {
	item := x.(itemWithValue)
	h.index[item.Tid] = len(h.items)
	h.items = append(h.items, item)
}

func (h *valueHeap) Pop() interface{} {
	n := len(h.items)
	item := h.items[n-1]
	h.items = h.items[:n-1]
	delete(h.index, item.Tid)
	return item
}

// TopKByValue tracks either the top-K or bottom-K entries depending on direction.
type TopKByValue struct {
	mu    sync.RWMutex
	k     int
	dir   Direction
	items map[int64]float64
	h     *valueHeap
	ttl   time.Duration
}

// NewTopKByValue creates a new TopKByValue for K elements, TTL, and direction.
func NewTopKByValue(k int, ttl time.Duration, dir Direction) *TopKByValue {
	h := &valueHeap{
		items: make([]itemWithValue, 0, k),
		index: make(map[int64]int),
		dir:   dir,
	}
	return &TopKByValue{
		k:     k,
		dir:   dir,
		items: make(map[int64]float64, k),
		h:     h,
		ttl:   ttl,
	}
}

// Add inserts or updates a tid with value, returns true if the set changed.
func (t *TopKByValue) Add(tid int64, value float64) bool {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if already tracked
	if existing, ok := t.items[tid]; ok {
		// For UP, only update if value > existing; for DOWN, if value < existing
		if (t.dir == Direction_UP && value <= existing) ||
			(t.dir == Direction_DOWN && value >= existing) {
			return true
		}

		// Update existing
		t.items[tid] = value
		if idx, ok := t.h.index[tid]; ok {
			t.h.items[idx].Value = value
			t.h.items[idx].LastSeen = now
			heap.Fix(t.h, idx)
			return true
		}
		// Fallback if index missing
		t.rebuildIndex()
	}

	// If under capacity, just push
	if len(t.h.items) < t.k {
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}

	// At capacity: compare against root
	rootValue := t.h.items[0].Value
	if (t.dir == Direction_UP && value > rootValue) ||
		(t.dir == Direction_DOWN && value < rootValue) {
		evicted := heap.Pop(t.h).(itemWithValue)
		delete(t.items, evicted.Tid)
		heap.Push(t.h, itemWithValue{Tid: tid, Value: value, LastSeen: now})
		t.items[tid] = value
		return true
	}

	return false
}

// rebuildIndex reconstructs h.index from items.
func (t *TopKByValue) rebuildIndex() {
	t.h.index = make(map[int64]int, len(t.h.items))
	for i, item := range t.h.items {
		t.h.index[item.Tid] = i
	}
}

// Eligible returns true if value would enter the heap.
func (t *TopKByValue) Eligible(value float64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.h.items) < t.k {
		return true
	}
	rootValue := t.h.items[0].Value
	if t.dir == Direction_UP {
		return value > rootValue
	}
	return value < rootValue
}

// CleanupExpired removes entries older than ttl.
func (t *TopKByValue) CleanupExpired() {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	j := 0
	for i := 0; i < len(t.h.items); i++ {
		it := t.h.items[i]
		if now.Sub(it.LastSeen) <= t.ttl {
			t.h.items[j] = it
			t.h.index[it.Tid] = j
			j++
		} else {
			delete(t.items, it.Tid)
		}
	}
	t.h.items = t.h.items[:j]
	if len(t.h.items) > 0 {
		heap.Init(t.h)
	} else {
		t.h.index = make(map[int64]int)
	}
}

// SortedSlice returns the items in final order: for UP descending, for DOWN ascending.
func (t *TopKByValue) SortedSlice() []itemWithValue {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]itemWithValue, len(t.h.items))
	copy(out, t.h.items)
	sort.Slice(out, func(i, j int) bool {
		if t.dir == Direction_UP {
			return out[i].Value > out[j].Value
		}
		return out[i].Value < out[j].Value
	})
	return out
}
