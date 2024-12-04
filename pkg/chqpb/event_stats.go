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
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/patrickmn/go-cache"
	"strconv"
	"sync/atomic"
	"time"
)

type EventStatsCache struct {
	cache         *cache.Cache
	lastFlushed   atomic.Pointer[time.Time]
	flushInterval time.Duration
}

func NewEventStatsCache(flushInterval time.Duration) *EventStatsCache {
	c := &EventStatsCache{
		cache:         cache.New(1*time.Hour, 10*time.Minute),
		flushInterval: flushInterval,
	}
	now := time.Now()
	c.lastFlushed.Store(&now)
	return c
}

func (e *EventStatsCache) Record(stat *EventStats, now time.Time) ([]*EventStats, error) {
	truncatedHour := now.Truncate(time.Hour)
	id := fmt.Sprintf("%d", stat.EventStatsKey(truncatedHour))
	if w, found := e.cache.Get(id); found {
		existingStats := w.(*EventStats)
		stat.Count += existingStats.Count
		stat.Size += existingStats.Size
	} else {
		previousHourId := fmt.Sprintf("%d", stat.EventStatsKey(truncatedHour.Add(-1*time.Hour)))
		e.cache.Delete(previousHourId)
	}
	e.cache.Set(id, stat, 70*time.Minute)
	var shouldFlush = false

	lastFlushed := e.lastFlushed.Load()
	shouldFlush = time.Since(*lastFlushed) > e.flushInterval
	if shouldFlush {
		var flushList []*EventStats
		for _, v := range e.cache.Items() {
			eventStats := v.Object.(*EventStats)
			flushList = append(flushList, eventStats)
		}
		now := time.Now()
		e.lastFlushed.Store(&now)
		return flushList, nil
	}
	return nil, nil
}

func (e *EventStats) EventStatsKey(tsHour time.Time) uint64 {
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
		fqn := fmt.Sprintf("%s.%s=%s", context, tagName, tagValue)
		hash.WriteString(fqn)
	}
	return hash.Sum64()
}
