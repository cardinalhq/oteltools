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
	"math"
	"sync"
	"time"

	"github.com/apache/datasketches-go/hll"
	"github.com/cespare/xxhash/v2"
)

type MetricStatsCache struct {
	hllCache      map[uint64]*MetricStatsWrapper
	itemsByHour   map[time.Time]map[uint64]bool
	lastFlushed   time.Time
	flushInterval time.Duration
	mu            sync.Mutex
}

func NewMetricStatsCache(flushInterval time.Duration) *MetricStatsCache {
	c := &MetricStatsCache{
		hllCache:      make(map[uint64]*MetricStatsWrapper),
		itemsByHour:   make(map[time.Time]map[uint64]bool),
		flushInterval: flushInterval,
	}
	c.lastFlushed = time.Now()
	return c
}

func (m *MetricStatsCache) Record(stat *MetricStats, tagValue string, now time.Time) ([]*MetricStats, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	truncatedHour := now.Truncate(time.Hour)
	previousHour := truncatedHour.Add(-1 * time.Hour)

	id := stat.Key(truncatedHour)
	if wrapper, found := m.hllCache[id]; found {
		currentEstimate, err := wrapper.GetEstimate()
		if err != nil {
			return nil, err
		}
		if tagValue == "" && len(stat.Hll) > 0 {
			err = wrapper.MergeWith(stat.Hll)
		} else {
			err = wrapper.Hll.UpdateString(tagValue)
		}
		if err != nil {
			return nil, err
		}
		newEstimate, err := wrapper.GetEstimate()
		if err != nil {
			return nil, err
		}
		wrapper.Dirty = math.Abs(newEstimate-currentEstimate) > 0.1
	} else {
		previousHourId := stat.Key(previousHour)
		if previousHourItems, ok := m.itemsByHour[previousHour]; ok {
			if _, exists := previousHourItems[previousHourId]; exists {
				delete(m.hllCache, previousHourId)
				delete(previousHourItems, previousHourId)
			}
		}

		sketch, err := hll.NewUnion(12)
		if err != nil {
			return nil, err
		}
		wrapper = &MetricStatsWrapper{
			Stats: stat,
			Hll:   sketch,
			Dirty: true,
		}
		if tagValue == "" && len(stat.Hll) > 0 {
			err = wrapper.MergeWith(stat.Hll)
			if err != nil {
				return nil, err
			}
		} else {
			err = wrapper.Hll.UpdateString(tagValue)
			if err != nil {
				return nil, err
			}
		}
		m.hllCache[id] = wrapper
		if _, ok := m.itemsByHour[truncatedHour]; !ok {
			m.itemsByHour[truncatedHour] = make(map[uint64]bool)
		}
		m.itemsByHour[truncatedHour][id] = true
	}
	shouldFlush := time.Since(m.lastFlushed) > m.flushInterval

	if shouldFlush {
		var flushList []*MetricStats
		for _, wrapper := range m.hllCache {
			if wrapper.Dirty {
				estimate, err := wrapper.GetEstimate()
				if err != nil {
					return nil, err
				}
				wrapper.Stats.CardinalityEstimate = estimate
				bytes, err := wrapper.Hll.ToCompactSlice()
				if err != nil {
					return nil, err
				}
				wrapper.Stats.Hll = bytes
				wrapper.Stats.TsHour = truncatedHour.UnixMilli()
				flushList = append(flushList, wrapper.Stats)
				wrapper.Dirty = false
			}
		}

		if len(flushList) > 0 {
			m.cleanupPreviousHour(previousHour)
			m.lastFlushed = now
			return flushList, nil
		}
	}
	return nil, nil
}

func (m *MetricStatsCache) cleanupPreviousHour(previousHour time.Time) {
	for hour, items := range m.itemsByHour {
		if hour.Before(previousHour) {
			for key := range items {
				delete(m.hllCache, key)
			}
			delete(m.itemsByHour, hour)
		}
	}
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
