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
	"github.com/apache/datasketches-go/hll"
	"github.com/cespare/xxhash/v2"
	"github.com/patrickmn/go-cache"
	"math"
	"sync/atomic"
	"time"
)

type MetricStatsCache struct {
	hllCache    *cache.Cache
	lastFlushed atomic.Pointer[time.Time]
}

func NewMetricStatsCache() *MetricStatsCache {
	c := &MetricStatsCache{
		hllCache: cache.New(1*time.Hour, 10*time.Minute),
	}
	now := time.Now()
	c.lastFlushed.Store(&now)
	return c
}

func (m *MetricStatsCache) Record(stat *MetricStats, tagValue string, now time.Time) ([]*MetricStats, error) {
	truncatedHour := now.Truncate(time.Hour)
	id := m.computeId(stat, truncatedHour)
	if w, found := m.hllCache.Get(id); found {
		wrapper := w.(*MetricStatsWrapper)
		currentEstimate, err := wrapper.GetEstimate()
		if err != nil {
			return nil, err
		}
		err = wrapper.Hll.UpdateString(tagValue)
		if err != nil {
			return nil, err
		}
		newEstimate, err := wrapper.GetEstimate()
		if err != nil {
			return nil, err
		}
		wrapper.Dirty = math.Round(newEstimate) != math.Round(currentEstimate)
		m.hllCache.Set(id, wrapper, 70*time.Minute)
	} else {
		previousHourId := m.computeId(stat, truncatedHour.Add(-1*time.Hour))
		m.hllCache.Delete(previousHourId)

		sketch, err := hll.NewHllSketchWithDefault()
		if err != nil {
			return nil, err
		}
		wrapper := &MetricStatsWrapper{
			Stats: stat,
			Hll:   sketch,
			Dirty: true,
		}
		err = wrapper.Hll.UpdateString(tagValue)
		if err != nil {
			return nil, err
		}
		m.hllCache.Set(id, wrapper, 70*time.Minute)
	}
	var shouldFlush = false

	lastFlushed := m.lastFlushed.Load()
	shouldFlush = time.Since(*lastFlushed) > 5*time.Minute

	if shouldFlush {
		var flushList []*MetricStats

		for _, v := range m.hllCache.Items() {
			wrapper := v.Object.(*MetricStatsWrapper)
			if wrapper.Dirty {
				wrapper.Dirty = false
			}
			estimate, err := wrapper.GetEstimate()
			if err != nil {
				return nil, err
			}
			wrapper.Stats.CardinalityEstimate = estimate
			flushList = append(flushList, wrapper.Stats)
		}
		m.lastFlushed.Store(&now)
		return flushList, nil
	}
	return nil, nil
}

func (m *MetricStatsCache) computeId(stat *MetricStats, truncatedHour time.Time) string {
	hash := xxhash.New()
	hash.WriteString(truncatedHour.String())
	hash.WriteString(stat.MetricName)
	hash.WriteString(stat.MetricType)
	hash.WriteString(stat.TagScope)
	hash.WriteString(stat.TagName)
	hash.WriteString(stat.ServiceName)
	hash.WriteString(stat.Phase.String())
	hash.WriteString(stat.ProcessorId)
	hash.WriteString(stat.CollectorId)
	hash.WriteString(stat.CustomerId)
	return fmt.Sprintf("%d", hash.Sum64())
}
