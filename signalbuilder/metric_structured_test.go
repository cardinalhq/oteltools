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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestParseMetrics(t *testing.T) {
	yamlData := []byte(`
resource:
  service.name: test-service
  service.version: 1.0.0
scopes:
  - name: test-scope
    version: 1.0.0
    attributes:
      scope.key: scope_value
    metrics:
      - name: test_gauge
        description: A test gauge metric
        unit: "ms"
        type: gauge
        gauge:
          data_points:
            - timestamp: 1609459200000000000
              value: 42.5
              attributes:
                label1: value1
      - name: test_sum
        description: A test sum metric
        unit: "bytes"
        type: sum
        sum:
          aggregation_temporality: cumulative
          is_monotonic: true
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              value: 100.0
              attributes:
                label2: value2
`)

	rm, err := ParseMetrics(yamlData)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Verify resource
	assert.Equal(t, "test-service", rm.Resource["service.name"])
	assert.Equal(t, "1.0.0", rm.Resource["service.version"])

	// Verify scopes
	require.Len(t, rm.ScopeMetrics, 1)
	scope := rm.ScopeMetrics[0]
	assert.Equal(t, "test-scope", scope.Name)
	assert.Equal(t, "1.0.0", scope.Version)
	assert.Equal(t, "scope_value", scope.Attributes["scope.key"])

	// Verify metrics
	require.Len(t, scope.Metrics, 2)

	// Verify gauge metric
	gauge := scope.Metrics[0]
	assert.Equal(t, "test_gauge", gauge.Name)
	assert.Equal(t, "A test gauge metric", gauge.Description)
	assert.Equal(t, "ms", gauge.Unit)
	assert.Equal(t, "gauge", gauge.Type)
	require.NotNil(t, gauge.Gauge)
	require.Len(t, gauge.Gauge.DataPoints, 1)
	
	gaugeDP := gauge.Gauge.DataPoints[0]
	assert.Equal(t, int64(1609459200000000000), gaugeDP.Timestamp)
	assert.Equal(t, 42.5, gaugeDP.Value)
	assert.Equal(t, "value1", gaugeDP.Attributes["label1"])

	// Verify sum metric
	sum := scope.Metrics[1]
	assert.Equal(t, "test_sum", sum.Name)
	assert.Equal(t, "A test sum metric", sum.Description)
	assert.Equal(t, "bytes", sum.Unit)
	assert.Equal(t, "sum", sum.Type)
	require.NotNil(t, sum.Sum)
	assert.Equal(t, "cumulative", sum.Sum.AggregationTemporality)
	assert.True(t, sum.Sum.IsMonotonic)
	require.Len(t, sum.Sum.DataPoints, 1)

	sumDP := sum.Sum.DataPoints[0]
	assert.Equal(t, int64(1609459200000000000), sumDP.Timestamp)
	assert.Equal(t, int64(1609459100000000000), sumDP.StartTimestamp)
	assert.Equal(t, 100.0, sumDP.Value)
	assert.Equal(t, "value2", sumDP.Attributes["label2"])
}

func TestParseMetricsJSON(t *testing.T) {
	jsonData := []byte(`{
  "resource": {
    "service.name": "json-service"
  },
  "scopes": [
    {
      "name": "json-scope",
      "metrics": [
        {
          "name": "json_gauge",
          "type": "gauge",
          "gauge": {
            "data_points": [
              {
                "timestamp": 1609459200000000000,
                "value": 123.45
              }
            ]
          }
        }
      ]
    }
  ]
}`)

	rm, err := ParseMetrics(jsonData)
	require.NoError(t, err)
	assert.Equal(t, "json-service", rm.Resource["service.name"])
	assert.Equal(t, "json-scope", rm.ScopeMetrics[0].Name)
	assert.Equal(t, "json_gauge", rm.ScopeMetrics[0].Metrics[0].Name)
}

func TestMustParseMetrics(t *testing.T) {
	validYAML := []byte(`
resource: {}
scopes:
  - metrics:
      - name: test
        type: gauge
        gauge:
          data_points:
            - timestamp: 1609459200000000000
              value: 1.0
`)

	// Should not panic
	rm := MustParseMetrics(validYAML)
	assert.NotNil(t, rm)

	// Should panic
	assert.Panics(t, func() {
		MustParseMetrics([]byte("invalid yaml"))
	})
}

func TestValidateMetrics(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid gauge",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: gauge
        gauge:
          data_points:
            - timestamp: 1609459200000000000
              value: 1.0
`,
			wantErr: false,
		},
		{
			name: "valid sum",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: sum
        sum:
          data_points:
            - timestamp: 1609459200000000000
              value: 1.0
`,
			wantErr: false,
		},
		{
			name: "no scopes",
			yaml: `
resource: {}
scopes: []
`,
			wantErr: true,
			errMsg:  "at least one scopes entry is required",
		},
		{
			name: "no metrics",
			yaml: `
resource: {}
scopes:
  - metrics: []
`,
			wantErr: true,
			errMsg:  "at least one metric is required",
		},
		{
			name: "unsupported type",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: histogram
`,
			wantErr: true,
			errMsg:  "unsupported metric type 'histogram'",
		},
		{
			name: "gauge without gauge field",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: gauge
`,
			wantErr: true,
			errMsg:  "gauge field is required when type is 'gauge'",
		},
		{
			name: "sum without sum field",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: sum
`,
			wantErr: true,
			errMsg:  "sum field is required when type is 'sum'",
		},
		{
			name: "gauge without data points",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: gauge
        gauge:
          data_points: []
`,
			wantErr: true,
			errMsg:  "at least one data point is required for gauge",
		},
		{
			name: "sum without data points",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: sum
        sum:
          data_points: []
`,
			wantErr: true,
			errMsg:  "at least one data point is required for sum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseMetrics([]byte(tt.yaml))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsBuilderAddFromYAML(t *testing.T) {
	yamlData := []byte(`
resource:
  service.name: test-service
scopes:
  - name: test-scope
    metrics:
      - name: test_gauge
        description: A test gauge
        unit: count
        type: gauge
        gauge:
          data_points:
            - timestamp: 1609459200000000000
              value: 42.0
              attributes:
                env: production
      - name: test_sum
        description: A test sum
        unit: bytes
        type: sum
        sum:
          aggregation_temporality: delta
          is_monotonic: false
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              value: 100.0
`)

	builder := NewMetricsBuilder()
	err := builder.AddFromYAML(yamlData)
	require.NoError(t, err)

	metrics := builder.Build()
	assert.Equal(t, 1, metrics.ResourceMetrics().Len())

	rm := metrics.ResourceMetrics().At(0)
	serviceNameVal, exists := rm.Resource().Attributes().Get("service.name")
	assert.True(t, exists)
	assert.Equal(t, "test-service", serviceNameVal.Str())

	assert.Equal(t, 1, rm.ScopeMetrics().Len())
	sm := rm.ScopeMetrics().At(0)
	assert.Equal(t, "test-scope", sm.Scope().Name())
	assert.Equal(t, 2, sm.Metrics().Len())

	// Verify gauge metric
	gaugeMetric := sm.Metrics().At(0)
	assert.Equal(t, "test_gauge", gaugeMetric.Name())
	assert.Equal(t, "A test gauge", gaugeMetric.Description())
	assert.Equal(t, "count", gaugeMetric.Unit())
	assert.Equal(t, pmetric.MetricTypeGauge, gaugeMetric.Type())
	assert.Equal(t, 1, gaugeMetric.Gauge().DataPoints().Len())

	gaugeDP := gaugeMetric.Gauge().DataPoints().At(0)
	assert.Equal(t, 42.0, gaugeDP.DoubleValue())
	envVal, exists := gaugeDP.Attributes().Get("env")
	assert.True(t, exists)
	assert.Equal(t, "production", envVal.Str())

	// Verify sum metric
	sumMetric := sm.Metrics().At(1)
	assert.Equal(t, "test_sum", sumMetric.Name())
	assert.Equal(t, "A test sum", sumMetric.Description())
	assert.Equal(t, "bytes", sumMetric.Unit())
	assert.Equal(t, pmetric.MetricTypeSum, sumMetric.Type())
	
	sum := sumMetric.Sum()
	assert.Equal(t, pmetric.AggregationTemporalityDelta, sum.AggregationTemporality())
	assert.False(t, sum.IsMonotonic())
	assert.Equal(t, 1, sum.DataPoints().Len())

	sumDP := sum.DataPoints().At(0)
	assert.Equal(t, 100.0, sumDP.DoubleValue())
}