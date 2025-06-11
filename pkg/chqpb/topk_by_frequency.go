// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"container/heap"
	"sort"
	"sync"
	"time"
)

type itemWithCount struct {
	Tid      int64
	Count    int64
	LastSeen time.Time
}

type freqMinHeap struct {
	items []itemWithCount
	index map[int64]int
}

func (h freqMinHeap) Len() int           { return len(h.items) }
func (h freqMinHeap) Less(i, j int) bool { return h.items[i].Count < h.items[j].Count }
func (h freqMinHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.index[h.items[i].Tid] = i
	h.index[h.items[j].Tid] = j
}

func (h *freqMinHeap) Push(x interface{}) {
	item := x.(itemWithCount)
	h.index[item.Tid] = len(h.items)
	h.items = append(h.items, item)
}

func (h *freqMinHeap) Pop() interface{} {
	n := len(h.items)
	item := h.items[n-1]
	h.items = h.items[0 : n-1]
	delete(h.index, item.Tid)
	return item
}

type TopKByFrequency struct {
	mu    sync.RWMutex
	k     int
	count map[int64]int64
	h     *freqMinHeap
	ttl   time.Duration
}

func NewTopKByFrequency(k int, ttl time.Duration) *TopKByFrequency {
	return &TopKByFrequency{
		k:     k,
		count: make(map[int64]int64),
		h:     &freqMinHeap{items: make([]itemWithCount, 0, k), index: make(map[int64]int)},
		ttl:   ttl,
	}
}

func (t *TopKByFrequency) AddCount(tid int64, count int) {
	if count <= 0 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.count[tid] += int64(count)
	now := time.Now()

	if i, ok := t.h.index[tid]; ok {
		t.h.items[i].Count = t.count[tid]
		t.h.items[i].LastSeen = now
		heap.Fix(t.h, i)
		return
	}

	if len(t.h.items) < t.k {
		heap.Push(t.h, itemWithCount{Tid: tid, Count: t.count[tid], LastSeen: now})
		return
	}

	if t.count[tid] > t.h.items[0].Count {
		heap.Pop(t.h)
		heap.Push(t.h, itemWithCount{Tid: tid, Count: t.count[tid], LastSeen: now})
	}
}

func (t *TopKByFrequency) Add(tid int64) {
	t.AddCount(tid, 1)
}

func (t *TopKByFrequency) Eligible(tid int64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.h.index[tid]
	return ok || len(t.h.items) == 0
}

func (t *TopKByFrequency) EligibleWithCount(tid int64, count int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	currentCount := t.count[tid] + int64(count)

	if len(t.h.items) < t.k {
		return true
	}
	if _, ok := t.h.index[tid]; ok {
		return true
	}
	if len(t.h.items) > 0 && currentCount > t.h.items[0].Count {
		return true
	}
	return false
}

// CleanupExpired removes entries from the heap whose LastSeen is older than TTL.
func (t *TopKByFrequency) CleanupExpired() {
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
			delete(t.count, item.Tid)
		}
	}
	t.h.items = t.h.items[:j]
	heap.Init(t.h)
}

func (t *TopKByFrequency) SortedSlice() []itemWithCount {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]itemWithCount, len(t.h.items))
	copy(out, t.h.items)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return out
}
