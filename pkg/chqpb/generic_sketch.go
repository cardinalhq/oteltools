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
}

func NewGenericSketchCache(interval time.Duration, cid, telemetryType string, flushFunc func(*GenericSketchList) error) *GenericSketchCache {
	c := &GenericSketchCache{
		interval:      interval,
		customerId:    cid,
		flushFunc:     flushFunc,
		telemetryType: telemetryType,
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

func (c *GenericSketchCache) UpdateWithCount(
	metricName string,
	metricType string,
	tagValues map[string]string,
	isAggregate bool,
	value float64,
	count uint64,
	ts time.Time,
) {
	interval := ts.Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map)

	val, ok := skMap.Load(tid)
	var entry *genericSketchEntry
	if !ok {
		m, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &genericSketchEntry{
			internal: m,
			proto: &GenericSketchProto{
				MetricName:  metricName,
				MetricType:  metricType,
				Tid:         tid,
				Interval:    interval,
				Tags:        tagValues,
				IsAggregate: isAggregate,
			},
		}
		skMap.Store(tid, entry)
	} else {
		entry = val.(*genericSketchEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.internal.AddWithCount(value, float64(count))
}

func (c *GenericSketchCache) Update(
	metricName string,
	metricType string,
	tagValues map[string]string,
	isAggregate bool,
	value float64,
	ts time.Time,
) {
	interval := ts.Truncate(c.interval).Unix()
	tid := computeTID(metricName, tagValues)

	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	skMap := bucketIface.(*sync.Map)

	val, ok := skMap.Load(tid)
	var entry *genericSketchEntry
	if !ok {
		m, _ := ddsketch.NewDefaultDDSketch(0.01)
		entry = &genericSketchEntry{
			internal: m,
			proto: &GenericSketchProto{
				MetricName:  metricName,
				MetricType:  metricType,
				Tid:         tid,
				Interval:    interval,
				Tags:        tagValues,
				IsAggregate: isAggregate,
			},
		}
		skMap.Store(tid, entry)
	} else {
		entry = val.(*genericSketchEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.internal.Add(value)
}

func (c *GenericSketchCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	list := &GenericSketchList{CustomerId: c.customerId, TelemetryType: c.telemetryType}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}

		skMap := v.(*sync.Map)
		skMap.Range(func(tid, val interface{}) bool {
			entry := val.(*genericSketchEntry)

			entry.mu.Lock()
			entry.proto.Sketch = Encode(entry.internal)
			list.Sketches = append(list.Sketches, entry.proto)
			entry.mu.Unlock()

			return true
		})
		c.buckets.Delete(intervalKey)
		return true
	})

	if len(list.Sketches) > 0 {
		if err := c.flushFunc(list); err != nil {
			slog.Error("failed to flush generic sketches", slog.String("customerId", c.customerId), slog.String("error", err.Error()))
		}
	}
}
