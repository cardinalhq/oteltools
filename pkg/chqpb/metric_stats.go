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
	"math"
	"strings"
	"time"
)

type MetricStatsCache struct {
	statsCache *StatsCache[*MetricStatsWrapper]
}

func initializeMetricStats(key string) (*MetricStatsWrapper, error) {
	wrapper := &MetricStatsWrapper{}
	wrapper.Stats = &MetricStats{}
	union, err := hll.NewUnion(12)
	if err != nil {
		return nil, err
	}
	wrapper.Hll = union
	return wrapper, nil
}

func NewMetricStatsCache(capacity int,
	numBins uint16,
	flushInterval time.Duration,
	flushCallback FlushCallback[*MetricStatsWrapper],
	initializeCallback InitializeCallback[*MetricStatsWrapper],
	clock Clock) *MetricStatsCache {
	c := &MetricStatsCache{
		statsCache: NewStatsCache[*MetricStatsWrapper](capacity, numBins, flushInterval, flushCallback, initializeCallback, clock),
	}
	return c
}

func updateMetricStats(phase Phase, existing *MetricStatsWrapper, metricName, metricType, tagScope, tagName, processorId, customerId, collectorId, tagValue string, attributes []*Attribute) error {
	existing.Stats.MetricName = metricName
	existing.Stats.MetricType = metricType
	existing.Stats.TagScope = tagScope
	existing.Stats.TagName = tagName
	existing.Stats.Phase = phase
	existing.Stats.ProcessorId = processorId
	existing.Stats.CustomerId = customerId
	existing.Stats.CollectorId = collectorId
	existing.Stats.Attributes = attributes
	previousEstimate, err := existing.GetEstimate()
	if err != nil {
		return err
	}
	err = existing.Hll.UpdateString(tagValue)
	if err != nil {
		return err
	}
	newEstimate, err := existing.GetEstimate()
	if err != nil {
		return err
	}
	existing.Dirty = math.Abs(newEstimate-previousEstimate) > 0.1
	return nil
}

func (e *MetricStatsCache) Record(phase Phase, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId, tagValue string, attributes []*Attribute) error {
	now := time.Now()
	truncatedHour := now.Truncate(time.Hour).UnixMilli()
	key := constructMetricStatsKey(phase, truncatedHour, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId, attributes)
	err := e.statsCache.Compute(key, func(existing *MetricStatsWrapper) error {
		err := updateMetricStats(phase, existing, metricName, metricType, tagScope, tagName, processorId, customerId, collectorId, tagValue, attributes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (e *MetricStatsCache) RecordMetricStats(metricStats *MetricStats) error {
	err := e.statsCache.Compute(metricStats.Key(), func(existing *MetricStatsWrapper) error {
		existing.Stats.MetricName = metricStats.MetricName
		existing.Stats.MetricType = metricStats.MetricType
		existing.Stats.TagScope = metricStats.TagScope
		existing.Stats.TagName = metricStats.TagName
		existing.Stats.ProcessorId = metricStats.ProcessorId
		existing.Stats.CustomerId = metricStats.CustomerId
		existing.Stats.CollectorId = metricStats.CollectorId
		existing.Stats.Attributes = metricStats.Attributes
		existing.Stats.TsHour = metricStats.TsHour
		previousEstimate, err := existing.GetEstimate()
		if err != nil {
			return err
		}
		err = existing.MergeWith(metricStats.Hll)
		if err != nil {
			return err
		}
		newEstimate, err := existing.GetEstimate()
		if err != nil {
			return err
		}
		existing.Dirty = math.Abs(newEstimate-previousEstimate) > 0.1
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *MetricStats) Key() string {
	return constructMetricStatsKey(
		m.Phase,
		m.TsHour,
		m.MetricName,
		m.MetricType,
		m.TagScope,
		m.TagName,
		m.ServiceName,
		m.ProcessorId,
		m.CollectorId,
		m.CustomerId,
		m.Attributes)
}

func constructMetricStatsKey(phase Phase, truncatedHour int64, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId string, attributes []*Attribute) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d:%s:%s:%s:%s:%s:%s:%s:%s:%s", truncatedHour, metricName, metricType, tagScope, tagName, serviceName, phase.String(), processorId, collectorId, customerId))
	for _, k := range attributes {
		sb.WriteString(fmt.Sprintf(":%s.%s=%s", k.ContextId, k.Key, k.Value))
	}
	return sb.String()
}
