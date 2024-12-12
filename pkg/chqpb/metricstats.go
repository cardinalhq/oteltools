// Copyright 2024 CardinalHQ, Inc
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
	"container/list"
	"github.com/apache/datasketches-go/hll"
	"github.com/cespare/xxhash/v2"
	"math"
	"sync"
	"time"
)

type MetricStatsCache struct {
	hllCache      map[uint64]*list.Element
	cacheOrder    *list.List
	cacheMutex    sync.RWMutex
	stopChan      chan struct{}
	cacheCapacity int
	flushInterval time.Duration
}

type LRUEntry struct {
	Key     uint64
	Wrapper *MetricStatsWrapper
}

func NewMetricStatsCache(cacheCapacity int, flushInterval time.Duration) *MetricStatsCache {
	c := &MetricStatsCache{
		hllCache:      make(map[uint64]*list.Element),
		cacheOrder:    list.New(),
		stopChan:      make(chan struct{}),
		cacheCapacity: cacheCapacity,
		flushInterval: flushInterval,
	}
	c.startCleanup()
	return c
}

func (m *MetricStatsCache) startCleanup() {
	go func() {
		ticker := time.NewTicker(m.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.cleanupStaleEntries()
			}
		}
	}()
}

func (m *MetricStatsCache) cleanupStaleEntries() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	now := time.Now()
	for e := m.cacheOrder.Back(); e != nil; {
		entry := e.Value.(*LRUEntry)
		if now.Sub(entry.Wrapper.lastUpdated) > time.Hour {
			delete(m.hllCache, entry.Key)
			next := e.Prev()
			m.cacheOrder.Remove(e)
			e = next
		} else {
			e = e.Prev()
		}
	}
}

func (m *MetricStatsCache) Record(stat *MetricStats, tagValue string, now time.Time) (*MetricStats, error) {
	id := stat.Key(now.Truncate(time.Hour))

	m.cacheMutex.RLock()
	elem, found := m.hllCache[id]
	m.cacheMutex.RUnlock()

	if found {
		entry := elem.Value.(*LRUEntry).Wrapper
		entry.entryMutex.Lock()
		defer entry.entryMutex.Unlock()

		return nil, m.updateWrapper(entry, stat, tagValue, now)
	}

	sketch, err := hll.NewUnion(12)
	if err != nil {
		return nil, err
	}
	wrapper := &MetricStatsWrapper{
		Stats:       stat,
		Hll:         sketch,
		lastUpdated: now,
	}

	err = m.updateWrapper(wrapper, stat, tagValue, now)
	if err != nil {
		return nil, err
	}

	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Double-check if the entry was added by another goroutine, when the cache level lock was being taken
	elem, found = m.hllCache[id]
	if found {
		entry := elem.Value.(*LRUEntry).Wrapper
		entry.entryMutex.Lock()
		defer entry.entryMutex.Unlock()
		return nil, m.updateWrapper(entry, stat, tagValue, now)
	}

	// Evict the oldest entry if over capacity
	if m.cacheOrder.Len() >= m.cacheCapacity {
		oldest := m.cacheOrder.Back()
		if oldest != nil {
			entry := oldest.Value.(*LRUEntry)
			delete(m.hllCache, entry.Key)
			m.cacheOrder.Remove(oldest)
		}
	}

	// Add the new entry to the cache
	elem = m.cacheOrder.PushFront(&LRUEntry{Key: id, Wrapper: wrapper})
	m.hllCache[id] = elem

	return stat, nil
}

func (m *MetricStatsCache) updateWrapper(wrapper *MetricStatsWrapper, stat *MetricStats, tagValue string, now time.Time) error {
	currentEstimate, err := wrapper.GetEstimate()
	if err != nil {
		return err
	}

	var mergeErr error
	if tagValue == "" && len(stat.Hll) > 0 {
		mergeErr = wrapper.MergeWith(stat.Hll)
	} else {
		mergeErr = wrapper.Hll.UpdateString(tagValue)
	}
	if mergeErr != nil {
		return mergeErr
	}

	newEstimate, err := wrapper.GetEstimate()
	if err != nil {
		return err
	}

	if math.Abs(newEstimate-currentEstimate) > 0.1 {
		wrapper.lastUpdated = now
		wrapper.Stats.Hll, _ = wrapper.Hll.ToCompactSlice()
		wrapper.Stats.CardinalityEstimate = newEstimate
	}
	return nil
}

func (m *MetricStats) Key(tsHour time.Time) uint64 {
	hash := xxhash.New()
	_, _ = hash.WriteString(tsHour.String())
	_, _ = hash.WriteString(m.MetricName)
	_, _ = hash.WriteString(m.MetricType)
	_, _ = hash.WriteString(m.TagScope)
	_, _ = hash.WriteString(m.TagName)
	_, _ = hash.WriteString(m.ServiceName)
	_, _ = hash.WriteString(m.Phase.String())
	_, _ = hash.WriteString(m.ProcessorId)
	_, _ = hash.WriteString(m.CollectorId)
	_, _ = hash.WriteString(m.CustomerId)
	for _, k := range m.Attributes {
		context := k.ContextId
		tagName := k.Key
		tagValue := k.Value
		_, _ = hash.WriteString(context)
		_, _ = hash.WriteString(".")
		_, _ = hash.WriteString(tagName)
		_, _ = hash.WriteString("=")
		_, _ = hash.WriteString(tagValue)
	}
	return hash.Sum64()
}
