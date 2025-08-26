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
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ResourceMetrics represents metrics grouped by resource
type ResourceMetrics struct {
	Resource     map[string]any `json:"resource,omitempty" yaml:"resource,omitempty"`
	ScopeMetrics []ScopeMetrics `json:"scopes" yaml:"scopes"`
}

// ScopeMetrics represents metrics grouped by instrumentation scope
type ScopeMetrics struct {
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
	Version    string         `json:"version,omitempty" yaml:"version,omitempty"`
	SchemaURL  string         `json:"schema_url,omitempty" yaml:"schema_url,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Metrics    []Metric       `json:"metrics" yaml:"metrics"`
}

// Metric represents a single metric
type Metric struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Unit        string         `json:"unit,omitempty" yaml:"unit,omitempty"`
	Type        string         `json:"type" yaml:"type"`
	Gauge       *GaugeMetric   `json:"gauge,omitempty" yaml:"gauge,omitempty"`
	Sum         *SumMetric     `json:"sum,omitempty" yaml:"sum,omitempty"`
	Summary     *SummaryMetric `json:"summary,omitempty" yaml:"summary,omitempty"`
}

// GaugeMetric represents a gauge metric
type GaugeMetric struct {
	DataPoints []NumberDataPoint `json:"data_points" yaml:"data_points"`
}

// SumMetric represents a sum metric
type SumMetric struct {
	AggregationTemporality string            `json:"aggregation_temporality,omitempty" yaml:"aggregation_temporality,omitempty"`
	IsMonotonic            bool              `json:"is_monotonic,omitempty" yaml:"is_monotonic,omitempty"`
	DataPoints             []NumberDataPoint `json:"data_points" yaml:"data_points"`
}

// SummaryMetric represents a summary metric
type SummaryMetric struct {
	DataPoints []SummaryDataPoint `json:"data_points" yaml:"data_points"`
}

// NumberDataPoint represents a number data point
type NumberDataPoint struct {
	Attributes     map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	StartTimestamp int64          `json:"start_timestamp,omitempty" yaml:"start_timestamp,omitempty"`
	Timestamp      int64          `json:"timestamp" yaml:"timestamp"`
	Value          float64        `json:"value" yaml:"value"`
	Flags          uint32         `json:"flags,omitempty" yaml:"flags,omitempty"`
}

// SummaryDataPoint represents a summary data point
type SummaryDataPoint struct {
	Attributes     map[string]any  `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	StartTimestamp int64           `json:"start_timestamp,omitempty" yaml:"start_timestamp,omitempty"`
	Timestamp      int64           `json:"timestamp" yaml:"timestamp"`
	Count          uint64          `json:"count" yaml:"count"`
	Sum            float64         `json:"sum" yaml:"sum"`
	Quantiles      []QuantileValue `json:"quantiles,omitempty" yaml:"quantiles,omitempty"`
	Flags          uint32          `json:"flags,omitempty" yaml:"flags,omitempty"`
}

// QuantileValue represents a quantile value in a summary
type QuantileValue struct {
	Quantile float64 `json:"quantile" yaml:"quantile"`
	Value    float64 `json:"value" yaml:"value"`
}

// ParseMetrics parses YAML (or JSON) data into ResourceMetrics
func ParseMetrics(data []byte, opts ...ParseOptions) (*ResourceMetrics, error) {
	strictMode := true
	if len(opts) > 0 {
		strictMode = opts[0].StrictMode
	}

	var rm ResourceMetrics
	reader := bytes.NewReader(data)
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(strictMode)

	if err := decoder.Decode(&rm); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateMetrics(&rm); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &rm, nil
}

// MustParseMetrics parses YAML bytes into ResourceMetrics, panicking on error.
// This is useful for unit tests where parsing is expected to succeed.
func MustParseMetrics(data []byte, opts ...ParseOptions) *ResourceMetrics {
	rm, err := ParseMetrics(data, opts...)
	if err != nil {
		panic(fmt.Sprintf("MustParseMetrics failed: %v", err))
	}
	return rm
}

func validateMetrics(rm *ResourceMetrics) error {
	if len(rm.ScopeMetrics) == 0 {
		return fmt.Errorf("at least one scopes entry is required")
	}

	for i, sm := range rm.ScopeMetrics {
		if len(sm.Metrics) == 0 {
			return fmt.Errorf("scopes[%d]: at least one metric is required", i)
		}

		for j, m := range sm.Metrics {
			if err := validateMetric(&m, fmt.Sprintf("scopes[%d].metrics[%d]", i, j)); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateMetric(m *Metric, prefix string) error {
	switch m.Type {
	case "gauge":
		if m.Gauge == nil {
			return fmt.Errorf("%s: gauge field is required when type is 'gauge'", prefix)
		}
		if len(m.Gauge.DataPoints) == 0 {
			return fmt.Errorf("%s: at least one data point is required for gauge", prefix)
		}
	case "sum":
		if m.Sum == nil {
			return fmt.Errorf("%s: sum field is required when type is 'sum'", prefix)
		}
		if len(m.Sum.DataPoints) == 0 {
			return fmt.Errorf("%s: at least one data point is required for sum", prefix)
		}
	case "summary":
		if m.Summary == nil {
			return fmt.Errorf("%s: summary field is required when type is 'summary'", prefix)
		}
		if len(m.Summary.DataPoints) == 0 {
			return fmt.Errorf("%s: at least one data point is required for summary", prefix)
		}
	default:
		return fmt.Errorf("%s: unsupported metric type '%s', only 'gauge', 'sum', and 'summary' are supported", prefix, m.Type)
	}
	return nil
}
