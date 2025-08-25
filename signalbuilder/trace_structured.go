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

// ResourceTraces represents traces grouped by resource
type ResourceTraces struct {
	Resource    map[string]any `json:"resource,omitempty" yaml:"resource,omitempty"`
	ScopeTraces []ScopeTraces  `json:"scopes" yaml:"scopes"`
}

// ScopeTraces represents traces grouped by instrumentation scope
type ScopeTraces struct {
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
	Version    string         `json:"version,omitempty" yaml:"version,omitempty"`
	SchemaURL  string         `json:"schema_url,omitempty" yaml:"schema_url,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Spans      []Span         `json:"spans" yaml:"spans"`
}

// Span represents a single span entry
type Span struct {
	TraceID                string         `json:"trace_id,omitempty" yaml:"trace_id,omitempty"`
	SpanID                 string         `json:"span_id,omitempty" yaml:"span_id,omitempty"`
	ParentSpanID           string         `json:"parent_span_id,omitempty" yaml:"parent_span_id,omitempty"`
	Name                   string         `json:"name" yaml:"name"`
	Kind                   int32          `json:"kind,omitempty" yaml:"kind,omitempty"`
	StartTimestamp         int64          `json:"start_timestamp,omitempty" yaml:"start_timestamp,omitempty"`
	EndTimestamp           int64          `json:"end_timestamp,omitempty" yaml:"end_timestamp,omitempty"`
	Attributes             map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	DroppedAttributesCount uint32         `json:"dropped_attributes_count,omitempty" yaml:"dropped_attributes_count,omitempty"`
	Events                 []SpanEvent    `json:"events,omitempty" yaml:"events,omitempty"`
	DroppedEventsCount     uint32         `json:"dropped_events_count,omitempty" yaml:"dropped_events_count,omitempty"`
	Links                  []SpanLink     `json:"links,omitempty" yaml:"links,omitempty"`
	DroppedLinksCount      uint32         `json:"dropped_links_count,omitempty" yaml:"dropped_links_count,omitempty"`
	Status                 SpanStatus     `json:"status,omitempty" yaml:"status,omitempty"`
	Flags                  uint32         `json:"flags,omitempty" yaml:"flags,omitempty"`
}

// SpanEvent represents a span event
type SpanEvent struct {
	Timestamp              int64          `json:"timestamp" yaml:"timestamp"`
	Name                   string         `json:"name" yaml:"name"`
	Attributes             map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	DroppedAttributesCount uint32         `json:"dropped_attributes_count,omitempty" yaml:"dropped_attributes_count,omitempty"`
}

// SpanLink represents a span link
type SpanLink struct {
	TraceID                string         `json:"trace_id" yaml:"trace_id"`
	SpanID                 string         `json:"span_id" yaml:"span_id"`
	TraceState             string         `json:"trace_state,omitempty" yaml:"trace_state,omitempty"`
	Attributes             map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	DroppedAttributesCount uint32         `json:"dropped_attributes_count,omitempty" yaml:"dropped_attributes_count,omitempty"`
}

// SpanStatus represents a span status
type SpanStatus struct {
	Code    int32  `json:"code,omitempty" yaml:"code,omitempty"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

// ParseTraces parses YAML (or JSON) data into ResourceTraces
func ParseTraces(data []byte, opts ...ParseOptions) (*ResourceTraces, error) {
	strictMode := true
	if len(opts) > 0 {
		strictMode = opts[0].StrictMode
	}

	var rt ResourceTraces
	reader := bytes.NewReader(data)
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(strictMode)

	if err := decoder.Decode(&rt); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateTraces(&rt); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &rt, nil
}

// MustParseTraces parses YAML bytes into ResourceTraces, panicking on error.
// This is useful for unit tests where parsing is expected to succeed.
func MustParseTraces(data []byte, opts ...ParseOptions) *ResourceTraces {
	rt, err := ParseTraces(data, opts...)
	if err != nil {
		panic(fmt.Sprintf("MustParseTraces failed: %v", err))
	}
	return rt
}

func validateTraces(rt *ResourceTraces) error {
	if len(rt.ScopeTraces) == 0 {
		return fmt.Errorf("at least one scopes entry is required")
	}

	for i, st := range rt.ScopeTraces {
		if len(st.Spans) == 0 {
			return fmt.Errorf("scopes[%d]: at least one span is required", i)
		}
	}

	return nil
}
