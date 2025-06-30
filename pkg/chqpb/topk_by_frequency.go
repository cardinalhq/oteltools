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

// Direction is your generated enum: UP vs DOWN (0 = unspecified, 1 = UP, 2 = DOWN).

type itemWithCount struct {
	Tid      int64
	Count    int64
	LastSeen time.Time
}

type freqHeap struct {
	items []itemWithCount
	index map[int64]int
	dir   Direction
}

func (h freqHeap) Len() int { return len(h.items) }
func (h freqHeap) Less(i, j int) bool {
	if h.dir == Direction_UP {
		return h.items[i].Count < h.items[j].Count // min-heap for UP
	}
	return h.items[i].Count > h.items[j].Count // max-heap for DOWN
}
func (h freqHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i].Tid] = i
	h.index[h.items[j].Tid] = j
}
func (h *freqHeap) Push(x interface{}) {
	it := x.(itemWithCount)
	h.index[it.Tid] = len(h.items)
	h.items = append(h.items, it)
}
func (h *freqHeap) Pop() interface{} {
	n := len(h.items)
	it := h.items[n-1]
	h.items = h.items[:n-1]
	delete(h.index, it.Tid)
	return it
}

// TopKByFrequency maintains a heap of size K, either the top-K (UP) or bottom-K (DOWN).
type TopKByFrequency struct {
	mu    sync.RWMutex
	k     int
	dir   Direction
	count map[int64]int64 // always retains every seen tid's count
	h     *freqHeap
	ttl   time.Duration
}

func NewTopKByFrequency(k int, ttl time.Duration, dir Direction) *TopKByFrequency {
	h := &freqHeap{
		items: make([]itemWithCount, 0, k),
		index: make(map[int64]int),
		dir:   dir,
	}
	return &TopKByFrequency{
		k:     k,
		dir:   dir,
		count: make(map[int64]int64),
		h:     h,
		ttl:   ttl,
	}
}

func (t *TopKByFrequency) AddCount(tid int64, delta int) bool {
	if delta <= 0 {
		return false
	}
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	// bump the count
	t.count[tid] += int64(delta)
	current := t.count[tid]

	// if already in heap, just update/fix
	if idx, in := t.h.index[tid]; in {
		if idx < 0 || idx >= len(t.h.items) {
			t.rebuildIndex()
			idx, in = t.h.index[tid]
		}
		if in {
			t.h.items[idx].Count = current
			t.h.items[idx].LastSeen = now
			heap.Fix(t.h, idx)
			return true
		}
	}

	// under capacity → push
	if len(t.h.items) < t.k {
		heap.Push(t.h, itemWithCount{Tid: tid, Count: current, LastSeen: now})
		return true
	}

	// compare against root
	root := t.h.items[0]
	keep := false
	if t.dir == Direction_UP {
		keep = current > root.Count
	} else {
		keep = current < root.Count
	}
	if keep {
		heap.Pop(t.h) // evict root, but **do NOT** delete t.count[root.Tid]
		heap.Push(t.h, itemWithCount{Tid: tid, Count: current, LastSeen: now})
		return true
	}
	return false
}

func (t *TopKByFrequency) Add(tid int64) bool {
	return t.AddCount(tid, 1)
}

func (t *TopKByFrequency) EligibleWithCount(tid int64, delta int) bool {
	if delta <= 0 {
		return false
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	newCount := t.count[tid] + int64(delta)

	// under capacity ⇒ eligible
	if len(t.h.items) < t.k {
		return true
	}
	// already in heap ⇒ stays eligible
	if _, in := t.h.index[tid]; in {
		return true
	}
	// compare against root
	rootCount := t.h.items[0].Count
	if t.dir == Direction_UP {
		return newCount > rootCount
	}
	return newCount < rootCount
}

func (t *TopKByFrequency) Eligible(tid int64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count, seen := t.count[tid]
	rootCount := int64(0)
	if len(t.h.items) > 0 {
		rootCount = t.h.items[0].Count
	}

	if !seen {
		// new tid: under capacity ⇒ eligible, else compare zero
		if len(t.h.items) < t.k {
			return true
		}
		if t.dir == Direction_UP {
			return count > rootCount // count==0
		}
		return count < rootCount // 0 < rootCount for DOWN
	}

	// existing tid
	if len(t.h.items) < t.k {
		return true
	}
	if _, in := t.h.index[tid]; in {
		return true
	}
	if t.dir == Direction_UP {
		return count > rootCount
	}
	return count < rootCount
}

func (t *TopKByFrequency) rebuildIndex() {
	t.h.index = make(map[int64]int, len(t.h.items))
	for i, it := range t.h.items {
		t.h.index[it.Tid] = i
	}
}

func (t *TopKByFrequency) CleanupExpired() {
	now := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	j := 0
	newIdx := make(map[int64]int, len(t.h.items))
	for _, it := range t.h.items {
		if now.Sub(it.LastSeen) <= t.ttl {
			t.h.items[j] = it
			newIdx[it.Tid] = j
			j++
		} else {
			delete(t.count, it.Tid)
		}
	}
	t.h.items = t.h.items[:j]
	t.h.index = newIdx
	if len(t.h.items) > 0 {
		heap.Init(t.h)
	}
}

func (t *TopKByFrequency) SortedSlice() []itemWithCount {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]itemWithCount, len(t.h.items))
	copy(out, t.h.items)
	sort.Slice(out, func(i, j int) bool {
		if t.dir == Direction_UP {
			return out[i].Count > out[j].Count
		}
		return out[i].Count < out[j].Count
	})
	return out
}
