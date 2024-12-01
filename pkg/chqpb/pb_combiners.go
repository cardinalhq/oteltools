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
	"github.com/cardinalhq/oteltools/pkg/stats"
	"github.com/cespare/xxhash"
)

func (l *EventStats) Key() uint64 {
	key := fmt.Sprintf("%s:%d:%d:%s:%s:%s", l.ServiceName, l.Fingerprint, int32(l.Phase), l.ProcessorId, l.CustomerId, l.CollectorId)
	key = AppendTagsToKey(l.Attributes, key)
	return xxhash.Sum64String(key)
}

func (l *EventStats) Matches(other stats.StatsObject) bool {
	otherLogStats, ok := other.(*EventStats)
	if !ok {
		return false
	}
	return l.Key() == otherLogStats.Key()
}

func (l *EventStats) Increment(_ string, count int, size int64) error {
	l.Count += int64(count)
	l.Size += size
	return nil
}

func (l *EventStats) Initialize() error {
	return nil
}

func AppendTagsToKey(tags []*Attribute, key string) string {
	for _, k := range tags {
		context := k.ContextId
		tagName := k.Key
		fqn := fmt.Sprintf("%s.%s", context, tagName)
		key += fmt.Sprintf(":%s=%s", fqn, k.Value)
	}
	return key
}

type MetricStatsWrapper struct {
	Stats *MetricStats
	Hll   hll.HllSketch
}

func (m *MetricStatsWrapper) Key() uint64 {
	stats := m.Stats
	key := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s", stats.MetricName,
		stats.MetricType,
		stats.TagScope,
		stats.TagName,
		stats.ServiceName,
		stats.Phase.String(),
		stats.ProcessorId,
		stats.CollectorId,
		stats.CustomerId)
	key = AppendTagsToKey(stats.Attributes, key)
	return xxhash.Sum64String(key)
}

func (m *MetricStatsWrapper) Increment(tag string, count int, _ int64) error {
	if err := m.Hll.UpdateString(tag); err != nil {
		return err
	}
	m.Stats.Count += int64(count)
	return nil
}

func (m *MetricStatsWrapper) Initialize() error {
	hll, err := hll.NewHllSketchWithDefault()
	if err != nil {
		return err
	}
	m.Hll = hll
	return nil
}

func (m *MetricStatsWrapper) Matches(other stats.StatsObject) bool {
	_, ok := other.(*MetricStatsWrapper)
	if !ok {
		return false
	}
	return m.Key() == other.Key()
}
