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
      - name: test_summary
        description: A test summary metric
        unit: "seconds"
        type: summary
        summary:
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 50
              sum: 250.5
              quantiles:
                - quantile: 0.5
                  value: 4.5
                - quantile: 0.95
                  value: 9.2
              attributes:
                label3: value3
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
	require.Len(t, scope.Metrics, 3)

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

	// Verify summary metric
	summary := scope.Metrics[2]
	assert.Equal(t, "test_summary", summary.Name)
	assert.Equal(t, "A test summary metric", summary.Description)
	assert.Equal(t, "seconds", summary.Unit)
	assert.Equal(t, "summary", summary.Type)
	require.NotNil(t, summary.Summary)
	require.Len(t, summary.Summary.DataPoints, 1)

	summaryDP := summary.Summary.DataPoints[0]
	assert.Equal(t, int64(1609459200000000000), summaryDP.Timestamp)
	assert.Equal(t, int64(1609459100000000000), summaryDP.StartTimestamp)
	assert.Equal(t, uint64(50), summaryDP.Count)
	assert.Equal(t, 250.5, summaryDP.Sum)
	assert.Equal(t, "value3", summaryDP.Attributes["label3"])

	require.Len(t, summaryDP.Quantiles, 2)
	assert.Equal(t, 0.5, summaryDP.Quantiles[0].Quantile)
	assert.Equal(t, 4.5, summaryDP.Quantiles[0].Value)
	assert.Equal(t, 0.95, summaryDP.Quantiles[1].Quantile)
	assert.Equal(t, 9.2, summaryDP.Quantiles[1].Value)
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
			name: "valid summary",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: summary
        summary:
          data_points:
            - timestamp: 1609459200000000000
              count: 10
              sum: 50.0
`,
			wantErr: false,
		},
		{
			name: "valid histogram",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: histogram
        histogram:
          data_points:
            - timestamp: 1609459200000000000
              count: 100
              sum: 500.0
`,
			wantErr: false,
		},
		{
			name: "valid exponential histogram",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: exponential_histogram
        exponential_histogram:
          data_points:
            - timestamp: 1609459200000000000
              count: 100
              sum: 500.0
`,
			wantErr: false,
		},
		{
			name: "unsupported type",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: unsupported_type
`,
			wantErr: true,
			errMsg:  "unsupported metric type 'unsupported_type'",
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
		{
			name: "summary without summary field",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: summary
`,
			wantErr: true,
			errMsg:  "summary field is required when type is 'summary'",
		},
		{
			name: "summary without data points",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: summary
        summary:
          data_points: []
`,
			wantErr: true,
			errMsg:  "at least one data point is required for summary",
		},
		{
			name: "histogram without histogram field",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: histogram
`,
			wantErr: true,
			errMsg:  "histogram field is required when type is 'histogram'",
		},
		{
			name: "histogram without data points",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: histogram
        histogram:
          data_points: []
`,
			wantErr: true,
			errMsg:  "at least one data point is required for histogram",
		},
		{
			name: "exponential histogram without exponential_histogram field",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: exponential_histogram
`,
			wantErr: true,
			errMsg:  "exponential_histogram field is required when type is 'exponential_histogram'",
		},
		{
			name: "exponential histogram without data points",
			yaml: `
resource: {}
scopes:
  - metrics:
      - name: test
        type: exponential_histogram
        exponential_histogram:
          data_points: []
`,
			wantErr: true,
			errMsg:  "at least one data point is required for exponential_histogram",
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
      - name: test_summary
        description: A test summary
        unit: seconds
        type: summary
        summary:
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 25
              sum: 125.0
              quantiles:
                - quantile: 0.5
                  value: 5.0
                - quantile: 0.99
                  value: 10.0
              attributes:
                region: us-east-1
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
	assert.Equal(t, 3, sm.Metrics().Len())

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

	// Verify summary metric
	summaryMetric := sm.Metrics().At(2)
	assert.Equal(t, "test_summary", summaryMetric.Name())
	assert.Equal(t, "A test summary", summaryMetric.Description())
	assert.Equal(t, "seconds", summaryMetric.Unit())
	assert.Equal(t, pmetric.MetricTypeSummary, summaryMetric.Type())

	summary := summaryMetric.Summary()
	assert.Equal(t, 1, summary.DataPoints().Len())

	summaryDP := summary.DataPoints().At(0)
	assert.Equal(t, uint64(25), summaryDP.Count())
	assert.Equal(t, 125.0, summaryDP.Sum())
	assert.Equal(t, 2, summaryDP.QuantileValues().Len())

	regionVal, exists := summaryDP.Attributes().Get("region")
	assert.True(t, exists)
	assert.Equal(t, "us-east-1", regionVal.Str())

	q1 := summaryDP.QuantileValues().At(0)
	assert.Equal(t, 0.5, q1.Quantile())
	assert.Equal(t, 5.0, q1.Value())

	q2 := summaryDP.QuantileValues().At(1)
	assert.Equal(t, 0.99, q2.Quantile())
	assert.Equal(t, 10.0, q2.Value())
}

func TestParseHistogramMetrics(t *testing.T) {
	yamlData := []byte(`
resource:
  service.name: histogram-service
scopes:
  - name: histogram-scope
    metrics:
      - name: test_histogram
        description: A test histogram metric
        unit: seconds
        type: histogram
        histogram:
          aggregation_temporality: cumulative
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 1000
              sum: 5000.0
              min: 0.1
              max: 50.0
              bucket_counts: [10, 20, 50, 100, 200, 300, 200, 100, 20]
              explicit_bounds: [0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 40.0]
              attributes:
                method: GET
      - name: test_exponential_histogram
        description: A test exponential histogram metric
        unit: milliseconds
        type: exponential_histogram
        exponential_histogram:
          aggregation_temporality: delta
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 500
              sum: 2500.0
              scale: 2
              zero_count: 50
              positive_buckets:
                offset: 1
                bucket_counts: [10, 20, 30, 40, 50]
              negative_buckets:
                offset: -2
                bucket_counts: [5, 10, 15]
              attributes:
                endpoint: /api/v1
`)

	rm, err := ParseMetrics(yamlData)
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Verify resource
	assert.Equal(t, "histogram-service", rm.Resource["service.name"])

	// Verify scopes
	require.Len(t, rm.ScopeMetrics, 1)
	scope := rm.ScopeMetrics[0]
	assert.Equal(t, "histogram-scope", scope.Name)

	// Verify metrics
	require.Len(t, scope.Metrics, 2)

	// Verify histogram metric
	histogram := scope.Metrics[0]
	assert.Equal(t, "test_histogram", histogram.Name)
	assert.Equal(t, "A test histogram metric", histogram.Description)
	assert.Equal(t, "seconds", histogram.Unit)
	assert.Equal(t, "histogram", histogram.Type)
	require.NotNil(t, histogram.Histogram)
	assert.Equal(t, "cumulative", histogram.Histogram.AggregationTemporality)
	require.Len(t, histogram.Histogram.DataPoints, 1)

	histDP := histogram.Histogram.DataPoints[0]
	assert.Equal(t, int64(1609459200000000000), histDP.Timestamp)
	assert.Equal(t, int64(1609459100000000000), histDP.StartTimestamp)
	assert.Equal(t, uint64(1000), histDP.Count)
	require.NotNil(t, histDP.Sum)
	assert.Equal(t, 5000.0, *histDP.Sum)
	require.NotNil(t, histDP.Min)
	assert.Equal(t, 0.1, *histDP.Min)
	require.NotNil(t, histDP.Max)
	assert.Equal(t, 50.0, *histDP.Max)
	assert.Equal(t, "GET", histDP.Attributes["method"])

	expectedBucketCounts := []uint64{10, 20, 50, 100, 200, 300, 200, 100, 20}
	assert.Equal(t, expectedBucketCounts, histDP.BucketCounts)

	expectedBounds := []float64{0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 40.0}
	assert.Equal(t, expectedBounds, histDP.ExplicitBounds)

	// Verify exponential histogram metric
	expHistogram := scope.Metrics[1]
	assert.Equal(t, "test_exponential_histogram", expHistogram.Name)
	assert.Equal(t, "A test exponential histogram metric", expHistogram.Description)
	assert.Equal(t, "milliseconds", expHistogram.Unit)
	assert.Equal(t, "exponential_histogram", expHistogram.Type)
	require.NotNil(t, expHistogram.ExponentialHistogram)
	assert.Equal(t, "delta", expHistogram.ExponentialHistogram.AggregationTemporality)
	require.Len(t, expHistogram.ExponentialHistogram.DataPoints, 1)

	expHistDP := expHistogram.ExponentialHistogram.DataPoints[0]
	assert.Equal(t, int64(1609459200000000000), expHistDP.Timestamp)
	assert.Equal(t, int64(1609459100000000000), expHistDP.StartTimestamp)
	assert.Equal(t, uint64(500), expHistDP.Count)
	require.NotNil(t, expHistDP.Sum)
	assert.Equal(t, 2500.0, *expHistDP.Sum)
	assert.Equal(t, int32(2), expHistDP.Scale)
	assert.Equal(t, uint64(50), expHistDP.ZeroCount)
	assert.Equal(t, "/api/v1", expHistDP.Attributes["endpoint"])

	// Verify positive buckets
	require.NotNil(t, expHistDP.PositiveBuckets)
	assert.Equal(t, int32(1), expHistDP.PositiveBuckets.Offset)
	expectedPosBuckets := []uint64{10, 20, 30, 40, 50}
	assert.Equal(t, expectedPosBuckets, expHistDP.PositiveBuckets.BucketCounts)

	// Verify negative buckets
	require.NotNil(t, expHistDP.NegativeBuckets)
	assert.Equal(t, int32(-2), expHistDP.NegativeBuckets.Offset)
	expectedNegBuckets := []uint64{5, 10, 15}
	assert.Equal(t, expectedNegBuckets, expHistDP.NegativeBuckets.BucketCounts)
}

func TestHistogramMetricsBuilderAddFromYAML(t *testing.T) {
	yamlData := []byte(`
resource:
  service.name: test-service
scopes:
  - name: test-scope
    metrics:
      - name: test_histogram
        description: A test histogram
        unit: seconds
        type: histogram
        histogram:
          aggregation_temporality: cumulative
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 100
              sum: 500.0
              bucket_counts: [10, 20, 30, 40]
              explicit_bounds: [1.0, 5.0, 10.0]
              attributes:
                status: success
      - name: test_exp_histogram
        description: A test exponential histogram
        unit: milliseconds
        type: exponential_histogram
        exponential_histogram:
          aggregation_temporality: delta
          data_points:
            - timestamp: 1609459200000000000
              start_timestamp: 1609459100000000000
              count: 200
              sum: 1000.0
              scale: 1
              zero_count: 10
              positive_buckets:
                offset: 0
                bucket_counts: [5, 10, 15, 20]
              attributes:
                region: us-west-2
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

	// Verify histogram metric
	histogramMetric := sm.Metrics().At(0)
	assert.Equal(t, "test_histogram", histogramMetric.Name())
	assert.Equal(t, "A test histogram", histogramMetric.Description())
	assert.Equal(t, "seconds", histogramMetric.Unit())
	assert.Equal(t, pmetric.MetricTypeHistogram, histogramMetric.Type())
	
	histogram := histogramMetric.Histogram()
	assert.Equal(t, pmetric.AggregationTemporalityCumulative, histogram.AggregationTemporality())
	assert.Equal(t, 1, histogram.DataPoints().Len())

	histDP := histogram.DataPoints().At(0)
	assert.Equal(t, uint64(100), histDP.Count())
	assert.Equal(t, 500.0, histDP.Sum())
	assert.Equal(t, 4, histDP.BucketCounts().Len())
	assert.Equal(t, 3, histDP.ExplicitBounds().Len())
	
	statusVal, exists := histDP.Attributes().Get("status")
	assert.True(t, exists)
	assert.Equal(t, "success", statusVal.Str())

	// Verify exponential histogram metric
	expHistogramMetric := sm.Metrics().At(1)
	assert.Equal(t, "test_exp_histogram", expHistogramMetric.Name())
	assert.Equal(t, "A test exponential histogram", expHistogramMetric.Description())
	assert.Equal(t, "milliseconds", expHistogramMetric.Unit())
	assert.Equal(t, pmetric.MetricTypeExponentialHistogram, expHistogramMetric.Type())
	
	expHistogram := expHistogramMetric.ExponentialHistogram()
	assert.Equal(t, pmetric.AggregationTemporalityDelta, expHistogram.AggregationTemporality())
	assert.Equal(t, 1, expHistogram.DataPoints().Len())

	expHistDP := expHistogram.DataPoints().At(0)
	assert.Equal(t, uint64(200), expHistDP.Count())
	assert.Equal(t, 1000.0, expHistDP.Sum())
	assert.Equal(t, int32(1), expHistDP.Scale())
	assert.Equal(t, uint64(10), expHistDP.ZeroCount())
	
	regionVal, exists := expHistDP.Attributes().Get("region")
	assert.True(t, exists)
	assert.Equal(t, "us-west-2", regionVal.Str())

	positive := expHistDP.Positive()
	assert.Equal(t, int32(0), positive.Offset())
	assert.Equal(t, 4, positive.BucketCounts().Len())
}
