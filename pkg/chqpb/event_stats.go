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
	"strings"
	"time"
)

type EventStatsCache struct {
	statsCache *StatsCache[*EventStats]
}

func NewEventStatsCache(capacity int, flushInterval time.Duration, flushCallback FlushCallback[*EventStats], clock Clock) *EventStatsCache {
	c := &EventStatsCache{
		statsCache: NewStatsCache[*EventStats](capacity, flushInterval, flushCallback, initializeEventStats, clock),
	}
	return c
}

func initializeEventStats() (*EventStats, error) {
	return &EventStats{}, nil
}

func updateEventStats(existing *EventStats,
	fingerprint int64,
	phase Phase,
	tsHour int64,
	serviceName,
	processorId, customerId, collectorId string,
	attributes []*Attribute,
	count int64, size int64) error {
	existing.Phase = phase
	existing.ServiceName = serviceName
	existing.Fingerprint = fingerprint
	existing.ProcessorId = processorId
	existing.CustomerId = customerId
	existing.CollectorId = collectorId
	existing.Attributes = attributes
	existing.TsHour = tsHour
	existing.Count += count
	existing.Size += size
	return nil
}

func (e *EventStatsCache) Record(serviceName string,
	fingerprint int64,
	phase Phase,
	processorId string,
	collectorId string,
	customerId string,
	attributes []*Attribute,
	count int64,
	size int64) error {
	now := time.Now()
	truncatedHour := now.Truncate(time.Hour).UnixMilli()
	key := constructEventStatsKey(serviceName, fingerprint, phase, processorId, customerId, collectorId, truncatedHour, attributes)
	err := e.statsCache.Compute(key, func(existing *EventStats) error {
		err := updateEventStats(existing, fingerprint, phase, truncatedHour, serviceName, processorId, customerId, collectorId, attributes, count, size)
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

func (e *EventStatsCache) RecordEventStats(eventStats *EventStats) error {
	key := constructEventStatsKey(eventStats.ServiceName, eventStats.Fingerprint, eventStats.Phase, eventStats.ProcessorId, eventStats.CustomerId, eventStats.CollectorId, eventStats.TsHour, eventStats.Attributes)
	err := e.statsCache.Compute(key, func(existing *EventStats) error {
		updateStatsObject(existing, eventStats)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func updateStatsObject(existing *EventStats, eventStats *EventStats) {
	existing.ProcessorId = eventStats.ProcessorId
	existing.CustomerId = eventStats.CustomerId
	existing.CollectorId = eventStats.CollectorId
	existing.Attributes = eventStats.Attributes
	existing.ServiceName = eventStats.ServiceName
	existing.Fingerprint = eventStats.Fingerprint
	existing.Phase = eventStats.Phase
	existing.TsHour = eventStats.TsHour
	existing.Count += eventStats.Count
	existing.Size += eventStats.Size
}

func constructEventStatsKey(serviceName string, fingerprint int64, phase Phase, processorId, customerId, collectorId string, truncatedHour int64, attributes []*Attribute) string {
	var sb strings.Builder
	sb.WriteString(serviceName)
	sb.WriteString(fmt.Sprintf(":%d:%d:%s:%s:%s:%d", fingerprint, int32(phase), processorId, customerId, collectorId, truncatedHour))
	for _, k := range attributes {
		sb.WriteString(fmt.Sprintf(":%s.%s=%s", k.ContextId, k.Key, k.Value))
	}
	return sb.String()
}
