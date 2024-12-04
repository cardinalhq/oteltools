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
	"github.com/cespare/xxhash/v2"
	"strconv"
	"sync"
	"time"
)

type EventStatsCache struct {
	cache         map[uint64]*eventStatsWrapper
	itemsByHour   map[time.Time]map[uint64]bool
	lastFlushed   time.Time
	flushInterval time.Duration
	mu            sync.Mutex
}

type eventStatsWrapper struct {
	stats   *EventStats
	isDirty bool
}

func NewEventStatsCache(flushInterval time.Duration) *EventStatsCache {
	c := &EventStatsCache{
		cache:         make(map[uint64]*eventStatsWrapper),
		itemsByHour:   make(map[time.Time]map[uint64]bool),
		flushInterval: flushInterval,
	}
	now := time.Now()
	c.lastFlushed = now
	return c
}

func (e *EventStatsCache) Record(stat *EventStats, now time.Time) ([]*EventStats, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	truncatedHour := now.Truncate(time.Hour)
	previousHour := truncatedHour.Add(-1 * time.Hour)

	esw := &eventStatsWrapper{
		stats:   stat,
		isDirty: false,
	}
	id := esw.key(truncatedHour)
	if existingESW, found := e.cache[id]; found {
		newCount := stat.Count + existingESW.stats.Count
		newSize := stat.Size + existingESW.stats.Size
		if newCount != existingESW.stats.Count || newSize != existingESW.stats.Size {
			existingESW.stats.Size = newSize
			existingESW.stats.Count = newCount
			existingESW.isDirty = true
		}
	} else {
		previousHourId := esw.key(previousHour)
		if previousHourItems, ok := e.itemsByHour[previousHour]; ok {
			if _, exists := previousHourItems[previousHourId]; exists {
				delete(e.cache, previousHourId)
			}
		}
		esw.isDirty = true
		e.cache[id] = esw
		if _, ok := e.itemsByHour[truncatedHour]; !ok {
			e.itemsByHour[truncatedHour] = make(map[uint64]bool)
		}
		e.itemsByHour[truncatedHour][id] = true
	}

	shouldFlush := time.Since(e.lastFlushed) > e.flushInterval
	if shouldFlush {
		var flushList []*EventStats
		for _, v := range e.cache {
			if v.isDirty {
				flushList = append(flushList, v.stats)
				v.isDirty = false
			}
		}
		if previousHourItems, ok := e.itemsByHour[previousHour]; ok {
			for key := range previousHourItems {
				delete(e.cache, key)
			}
			delete(e.itemsByHour, previousHour)
		}
		e.lastFlushed = time.Now()
		return flushList, nil
	}
	return nil, nil
}

func (esw *eventStatsWrapper) key(tsHour time.Time) uint64 {
	e := esw.stats
	hash := xxhash.New()
	hash.WriteString(tsHour.String())
	hash.WriteString(e.ServiceName)
	hash.WriteString(strconv.FormatInt(e.Fingerprint, 10))
	hash.WriteString(e.Phase.String())
	hash.WriteString(e.ProcessorId)
	hash.WriteString(e.CollectorId)
	hash.WriteString(e.CustomerId)
	for _, k := range e.Attributes {
		context := k.ContextId
		tagName := k.Key
		tagValue := k.Value
		hash.WriteString(context)
		hash.WriteString(".")
		hash.WriteString(tagName)
		hash.WriteString("=")
		hash.WriteString(tagValue)
	}
	return hash.Sum64()
}
