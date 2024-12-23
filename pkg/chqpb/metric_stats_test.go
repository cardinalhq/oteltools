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
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricStatsCacheAggregation(t *testing.T) {
	mockClock := NewMockClock(time.Now())

	var wg sync.WaitGroup

	var flushedItems []*MetricStatsWrapper
	flushCallback := func(expiredItems []*MetricStatsWrapper) {
		flushedItems = append(flushedItems, expiredItems...)
		wg.Done()
	}

	capacity := 3
	expiry := 5 * time.Minute
	cache := NewMetricStatsCache(capacity, 1, expiry, flushCallback, InitializeMetricStats, mockClock)

	metricName := "metricA"
	metricType := "counter"
	tagScope := "scopeA"
	tagName := "tagA"
	serviceName := "serviceA"
	processorId := "proc1"
	collectorId := "coll1"
	customerId := "cust1"
	tagValue := "tagValue1"
	attributes := []*Attribute{
		{ContextId: "context1", Key: "key1", Value: "value1"},
	}

	err := cache.Record(Phase_PRE, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId, tagValue, attributes)
	assert.NoError(t, err)

	err = cache.Record(Phase_PRE, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId, tagValue, attributes)
	assert.NoError(t, err)

	err = cache.Record(Phase_PRE, metricName, metricType, tagScope, tagName, serviceName, processorId, collectorId, customerId, "tagValue2", attributes)
	assert.NoError(t, err)

	wg.Add(1)
	mockClock.Advance(expiry + 1*time.Second) // Advance the clock to trigger the expiration of the previous bucket

	// Trigger cleanup manually
	cache.statsCache.cleanupExpiredEntries()
	wg.Wait()

	// Verify the flushed items
	assert.Equal(t, 1, len(flushedItems), "Only one bucket should be flushed")
	flushed := flushedItems[0]

	assert.Equal(t, metricName, flushed.Stats.MetricName, "Metric name should match")
	assert.Equal(t, metricType, flushed.Stats.MetricType, "Metric type should match")
	assert.Equal(t, tagScope, flushed.Stats.TagScope, "Tag scope should match")
	assert.Equal(t, tagName, flushed.Stats.TagName, "Tag name should match")
	assert.Equal(t, processorId, flushed.Stats.ProcessorId, "Processor ID should match")
	assert.Equal(t, collectorId, flushed.Stats.CollectorId, "Collector ID should match")
	assert.Equal(t, customerId, flushed.Stats.CustomerId, "Customer ID should match")
	assert.Equal(t, attributes, flushed.Stats.Attributes, "Attributes should match")

	estimate, err := flushed.GetEstimate()
	assert.NoError(t, err)
	assert.Greater(t, estimate, float64(0), "HLL estimate should be greater than 0")
	assert.True(t, math.Abs(estimate-2) < 0.1, "HLL estimate should reflect the number of unique tagValues")

	assert.True(t, flushed.Dirty, "Dirty flag should be set to true")
}
