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

package signalbuilder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestMetricScopeBuilder_Metric(t *testing.T) {
	scope := pmetric.NewScopeMetrics()
	msb := NewMetricScopeBuilder(scope)

	t.Run("existing metric", func(t *testing.T) {
		name := "test_metric"
		// Create the metric first
		gb := msb.Gauge(name)
		assert.NotNil(t, gb)

		// Retrieve the existing metric
		item2 := msb.Gauge(name)
		assert.Equal(t, item2, item2)

		// Retrieve the same name with different units
		item3 := msb.Gauge("name_3")
		assert.NotEqual(t, item2, item3)
	})

	t.Run("new gauge metric", func(t *testing.T) {
		item := msb.Gauge("new_gauge_metric")
		assert.NotNil(t, item)
	})

	t.Run("new sum metric", func(t *testing.T) {
		item := msb.Sum("new_sum_metric")
		assert.NotNil(t, item)
	})
}

func TestMetricScopeBuilder_Get(t *testing.T) {
	scope := pmetric.NewScopeMetrics()
	msb := NewMetricScopeBuilder(scope)

	t.Run("get scope metrics", func(t *testing.T) {
		result := msb.Get()
		assert.Equal(t, scope, result)
	})
}
