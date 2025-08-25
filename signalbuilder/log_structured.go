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

	"go.opentelemetry.io/collector/pdata/pcommon"
	"gopkg.in/yaml.v3"
)

// ResourceLogs represents logs grouped by resource
type ResourceLogs struct {
	Resource  map[string]any `json:"resource,omitempty" yaml:"resource,omitempty"`
	ScopeLogs []ScopeLogs    `json:"scopes" yaml:"scopes"`
}

// ScopeLogs represents logs grouped by instrumentation scope
type ScopeLogs struct {
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
	Version    string         `json:"version,omitempty" yaml:"version,omitempty"`
	SchemaURL  string         `json:"schema_url,omitempty" yaml:"schema_url,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	LogRecords []LogRecord    `json:"records" yaml:"records"`
}

// LogRecord represents a single log entry
type LogRecord struct {
	Timestamp              int64          `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	ObservedTimestamp      int64          `json:"observed_timestamp,omitempty" yaml:"observed_timestamp,omitempty"`
	SeverityText           string         `json:"severity_text,omitempty" yaml:"severity_text,omitempty"`
	SeverityNumber         int32          `json:"severity_number,omitempty" yaml:"severity_number,omitempty"`
	Body                   any            `json:"body" yaml:"body"`
	Attributes             map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	TraceID                string         `json:"trace_id,omitempty" yaml:"trace_id,omitempty"`
	SpanID                 string         `json:"span_id,omitempty" yaml:"span_id,omitempty"`
	Flags                  uint32         `json:"flags,omitempty" yaml:"flags,omitempty"`
	DroppedAttributesCount uint32         `json:"dropped_attributes_count,omitempty" yaml:"dropped_attributes_count,omitempty"`
	EventName              string         `json:"event_name,omitempty" yaml:"event_name,omitempty"`
}

// ParseOptions controls parsing behavior
type ParseOptions struct {
	StrictMode bool // If true, fail on unknown fields (default: true)
}

// Parse parses YAML (or JSON) data into ResourceLogs
func Parse(data []byte, opts ...ParseOptions) (*ResourceLogs, error) {
	strictMode := true
	if len(opts) > 0 {
		strictMode = opts[0].StrictMode
	}

	var rl ResourceLogs
	reader := bytes.NewReader(data)
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(strictMode)

	if err := decoder.Decode(&rl); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&rl); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &rl, nil
}

// MustParse parses YAML bytes into ResourceLogs, panicking on error.
// This is useful for unit tests where parsing is expected to succeed.
func MustParse(data []byte, opts ...ParseOptions) *ResourceLogs {
	rl, err := Parse(data, opts...)
	if err != nil {
		panic(fmt.Sprintf("MustParse failed: %v", err))
	}
	return rl
}

// validate performs validation on the parsed ResourceLogs
func validate(rl *ResourceLogs) error {
	if len(rl.ScopeLogs) == 0 {
		return fmt.Errorf("at least one scopes entry is required")
	}

	for i, sl := range rl.ScopeLogs {
		if len(sl.LogRecords) == 0 {
			return fmt.Errorf("scopes[%d]: at least one record is required", i)
		}
	}

	return nil
}

// fromRaw converts a map[string]any to pcommon.Map
func fromRaw(attrs map[string]any) (pcommon.Map, error) {
	result := pcommon.NewMap()
	if attrs == nil {
		return result, nil
	}

	// Use FromRaw for direct conversion
	if err := result.FromRaw(attrs); err != nil {
		return pcommon.NewMap(), fmt.Errorf("failed to convert attributes: %w", err)
	}
	return result, nil
}
