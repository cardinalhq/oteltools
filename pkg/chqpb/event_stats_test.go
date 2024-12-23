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
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestAggregationAndFlush(t *testing.T) {
	mockClock := NewMockClock(time.Now())

	var wg sync.WaitGroup

	var flushedItems []*EventStats
	flushCallback := func(expiredItems []*EventStats) {
		flushedItems = append(flushedItems, expiredItems...)
		wg.Done()
	}

	capacity := 3
	expiry := 5 * time.Minute
	cache := NewEventStatsCache(capacity, 1, expiry, flushCallback, mockClock)

	serviceName := "serviceA"
	fingerprint := int64(123)
	phase := Phase_PRE
	processorId := "proc1"
	collectorId := "coll1"
	customerId := "cust1"
	attributes := []*Attribute{
		{ContextId: "context1", Key: "key1", Value: "value1"},
	}

	err := cache.Record(serviceName, fingerprint, phase, processorId, collectorId, customerId, attributes, 5, 100)
	assert.NoError(t, err)

	err = cache.Record(serviceName, fingerprint, phase, processorId, collectorId, customerId, attributes, 10, 200)
	assert.NoError(t, err)

	err = cache.Record(serviceName, fingerprint, phase, processorId, collectorId, customerId, attributes, 15, 300)
	assert.NoError(t, err)

	wg.Add(1)
	mockClock.Advance(expiry + 1*time.Second)

	cache.statsCache.cleanupExpiredEntries()

	wg.Wait()

	assert.Equal(t, 1, len(flushedItems), "Only one bucket should be flushed")
	flushed := flushedItems[0]
	assert.Equal(t, serviceName, flushed.ServiceName, "Service name should match")
	assert.Equal(t, fingerprint, flushed.Fingerprint, "Fingerprint should match")
	assert.Equal(t, phase, flushed.Phase, "Phase should match")
	assert.Equal(t, processorId, flushed.ProcessorId, "Processor ID should match")
	assert.Equal(t, collectorId, flushed.CollectorId, "Collector ID should match")
	assert.Equal(t, customerId, flushed.CustomerId, "Customer ID should match")
	assert.Equal(t, int64(30), flushed.Count, "Count should be aggregated")
	assert.Equal(t, int64(600), flushed.Size, "Size should be aggregated")

	cache.statsCache.Close()
}

func TestCapacityConstraints(t *testing.T) {
	var wg sync.WaitGroup

	var flushedItems []*EventStats
	flushCallback := func(expiredItems []*EventStats) {
		flushedItems = append(flushedItems, expiredItems...)
		fmt.Printf("Flush callback invoked: %v items flushed\n", len(expiredItems))
		wg.Done()
	}

	mockClock := NewMockClock(time.Now().Truncate(time.Hour))
	cache := NewEventStatsCache(3, 1, 5*time.Minute, flushCallback, mockClock)

	err := cache.Record("serviceA", 123, Phase_PRE, "proc1", "coll1", "cust1", nil, 5, 100)
	assert.NoError(t, err)

	err = cache.Record("serviceB", 456, Phase_PRE, "proc2", "coll2", "cust2", nil, 10, 200)
	assert.NoError(t, err)

	err = cache.Record("serviceC", 789, Phase_PRE, "proc3", "coll3", "cust3", nil, 20, 300)
	assert.NoError(t, err)

	wg.Add(1)
	err = cache.Record("serviceD", 101112, Phase_PRE, "proc4", "coll4", "cust4", nil, 15, 150)
	assert.NoError(t, err)

	wg.Wait()

	assert.Equal(t, 1, len(flushedItems), "Exactly one item should be evicted when capacity is exceeded")

	evictedKey := constructEventStatsKey("serviceD", 101112, Phase_PRE, "proc4", "cust4", "coll4", mockClock.Now().UnixMilli(), nil)
	for _, item := range flushedItems {
		assert.NotEqual(t, evictedKey, constructEventStatsKey(item.ServiceName, item.Fingerprint, item.Phase, item.ProcessorId, item.CustomerId, item.CollectorId, item.TsHour, item.Attributes), "The new entry should not be evicted")
	}

	cache.statsCache.Close()
}
