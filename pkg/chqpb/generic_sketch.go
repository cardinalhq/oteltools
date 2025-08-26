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

// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"sync"
	"time"

	"github.com/DataDog/sketches-go/ddsketch"
	"golang.org/x/exp/slog"
)

type genericSketchEntry struct {
	mu       sync.Mutex
	proto    *GenericSketchProto
	internal *ddsketch.DDSketch
}

type GenericSketchCache struct {
	buckets       sync.Map // map[int64]*sync.Map where inner map[string]*genericSketchEntry
	customerId    string
	interval      time.Duration
	telemetryType string
	flushFunc     func(*GenericSketchList) error
	valueTopKs    sync.Map
	maxK          int
}

func NewGenericSketchCache(interval time.Duration, cid, telemetryType string, k int, flushFunc func(*GenericSketchList) error) *GenericSketchCache {
	c := &GenericSketchCache{
		interval:      interval,
		customerId:    cid,
		flushFunc:     flushFunc,
		telemetryType: telemetryType,
		maxK:          k,
	}
	go c.loop()
	return c
}

func (c *GenericSketchCache) loop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.flush()
	}
}

const (
	Count = "count"
)

//func (c *GenericSketchCache) getFreqTopK(metricName string, direction Direction, parentTid int64, tagFamilyId int64) *TopKByFrequency {
//	key := getKey(metricName, direction, parentTid, tagFamilyId)
//	topK, _ := c.freqTopKs.LoadOrStore(key, NewTopKByFrequency(c.maxK, 2*c.interval, direction))
//	return topK.(*TopKByFrequency)
//}

func (c *GenericSketchCache) getValueTopK(metricName string, direction Direction, parentTid int64, tagFamilyId int64) *TopKByValue {
	key := getKey(metricName, direction, parentTid, tagFamilyId)
	topK, _ := c.valueTopKs.LoadOrStore(key, NewTopKByValue(c.maxK, 2*c.interval, direction))
	return topK.(*TopKByValue)
}

func (c *GenericSketchCache) UpdateWithCount(
	metricName string,
	metricType string,
	direction Direction,
	tagValues map[string]string,
	parentTID int64,
	tagFamilyID int64,
	value float64,
	count uint64,
	ts time.Time,
) int64 {
	interval := ts.Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map)

	val, ok := skMap.Load(tid)

	valueTopK := c.getValueTopK(metricName, direction, parentTID, tagFamilyID)
	shouldAdd := valueTopK.Eligible(value)

	var entry *genericSketchEntry
	if !ok {
		if !shouldAdd {
			return tid
		}
		m, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &genericSketchEntry{
			internal: m,
			proto: &GenericSketchProto{
				MetricName:  metricName,
				MetricType:  metricType,
				Tid:         tid,
				Interval:    interval,
				Tags:        tagValues,
				ParentTID:   parentTID,
				TagFamilyId: tagFamilyID,
				Direction:   direction,
			},
		}
		skMap.Store(tid, entry)
	} else {
		entry = val.(*genericSketchEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.internal.AddWithCount(value, float64(count))
	return tid
}

func (c *GenericSketchCache) Update(
	metricName string,
	metricType string,
	direction Direction,
	tagValues map[string]string,
	parentTID int64,
	tagFamilyId int64,
	value float64,
	ts time.Time,
) int64 {
	interval := ts.Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map)

	valueTopK := c.getValueTopK(metricName, direction, parentTID, tagFamilyId)
	shouldAdd := valueTopK.Eligible(value)

	val, ok := skMap.Load(tid)
	var entry *genericSketchEntry
	if !ok {
		if !shouldAdd {
			return tid
		}
		m, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &genericSketchEntry{
			internal: m,
			proto: &GenericSketchProto{
				MetricName:  metricName,
				MetricType:  metricType,
				Tid:         tid,
				Interval:    interval,
				Tags:        tagValues,
				ParentTID:   parentTID,
				TagFamilyId: tagFamilyId,
				Direction:   direction,
			},
		}
		skMap.Store(tid, entry)
	} else {
		entry = val.(*genericSketchEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.internal.Add(value)
	return tid
}

func (c *GenericSketchCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	out := &GenericSketchList{
		CustomerId:    c.customerId,
		TelemetryType: c.telemetryType,
	}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}
		skMap := v.(*sync.Map)

		// ─── Phase 1: Populate both heaps ─────────────────────────────────
		skMap.Range(func(_, val interface{}) bool {
			entry := val.(*genericSketchEntry)
			mn, tid := entry.proto.MetricName, entry.proto.Tid

			// 1a) Expire old entries
			//freqTK := c.getFreqTopK(mn, entry.proto.Direction, entry.proto.ParentTID, entry.proto.TagFamilyId)
			//freqTK.CleanupExpired()
			valTK := c.getValueTopK(mn, entry.proto.Direction, entry.proto.ParentTID, entry.proto.TagFamilyId)
			valTK.CleanupExpired()

			// 1b) Update with this TID’s final score for the interval
			//switch entry.proto.MetricType {
			//case Count:
			//	cnt := int(math.Ceil(entry.internal.GetCount()))
			//	freqTK.AddCount(tid, cnt)
			//default:
			//	if p50, err := entry.internal.GetValueAtQuantile(0.5); err == nil {
			//		valTK.Add(tid, p50)
			//	}
			//}
			if p50, err := entry.internal.GetValueAtQuantile(0.5); err == nil {
				valTK.Add(tid, p50)
			}
			return true
		})

		// ─── Phase 2: Emit *only* the winners ─────────────────────────────
		skMap.Range(func(_, val interface{}) bool {
			entry := val.(*genericSketchEntry)
			mn, tid := entry.proto.MetricName, entry.proto.Tid

			//freqTK := c.getFreqTopK(mn, entry.proto.Direction, entry.proto.ParentTID, entry.proto.TagFamilyId)
			valTK := c.getValueTopK(mn, entry.proto.Direction, entry.proto.ParentTID, entry.proto.TagFamilyId)

			//inFreq := false
			//if entry.proto.MetricType == Count {
			//	// if tid sits in the heap’s index, it’s one of the top-K
			//	_, inFreq = freqTK.h.index[tid]
			//}

			//inVal := false
			//if entry.proto.MetricType != Count {
			//	_, inVal = valTK.h.index[tid]
			//}
			_, inVal := valTK.h.index[tid]

			//if !inFreq && !inVal {
			//	// neither top-K by count nor by value → skip
			//	return true
			//}
			if !inVal {
				// neither top-K by count nor by value → skip
				return true
			}

			// serialize & collect
			entry.mu.Lock()
			entry.proto.Sketch = Encode(entry.internal)
			out.Sketches = append(out.Sketches, entry.proto)
			entry.mu.Unlock()
			return true
		})

		c.buckets.Delete(intervalKey)
		return true
	})

	if len(out.Sketches) > 0 {
		if err := c.flushFunc(out); err != nil {
			slog.Error("failed to flush generic sketches",
				slog.String("customerId", c.customerId),
				slog.String("error", err.Error()),
			)
		}
	}
}
