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
	id := fmt.Sprintf("%d", stat.Key(truncatedHour))
	if w, found := m.hllCache.Get(id); found {
		wrapper := w.(*MetricStatsWrapper)
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
		wrapper.Dirty = math.Round(newEstimate) != math.Round(currentEstimate)
		m.hllCache.Set(id, wrapper, 70*time.Minute)
	} else {
		previousHourId := fmt.Sprintf("%d", stat.Key(truncatedHour.Add(-1*time.Hour)))
		m.hllCache.Delete(previousHourId)

		sketch, err := hll.NewUnion(12)
		if err != nil {
			return nil, err
		}
		wrapper := &MetricStatsWrapper{
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
			bytes, err := wrapper.Hll.ToCompactSlice()
			if err != nil {
				return nil, err
			}
			wrapper.Stats.Hll = bytes
			wrapper.Stats.TsHour = truncatedHour.UnixMilli()
			flushList = append(flushList, wrapper.Stats)
		}
		m.lastFlushed.Store(&now)
		return flushList, nil
	}
	return nil, nil
}

func (m *MetricStats) Key(tsHour time.Time) uint64 {
	hash := xxhash.New()
	hash.WriteString(tsHour.String())
	hash.WriteString(m.MetricName)
	hash.WriteString(m.MetricType)
	hash.WriteString(m.TagScope)
	hash.WriteString(m.TagName)
	hash.WriteString(m.ServiceName)
	hash.WriteString(m.Phase.String())
	hash.WriteString(m.ProcessorId)
	hash.WriteString(m.CollectorId)
	hash.WriteString(m.CustomerId)
	return hash.Sum64()
}
