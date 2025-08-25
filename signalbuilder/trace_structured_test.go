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
	"encoding/hex"
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	"github.com/stretchr/testify/assert"
)

func TestParseTraces(t *testing.T) {
	yamlData := `
resource:
  service.name: "my-service"
  service.version: "1.0.0"
scopes:
  - name: "my.tracer"
    version: "1.0"
    schema_url: "https://opentelemetry.io/schemas/1.30.0"
    attributes:
      library.name: "my-library"
    spans:
      - name: "test-span"
        trace_id: "12345678901234567890123456789012"
        span_id: "1234567890123456"
        parent_span_id: "9876543210987654"
        kind: 1
        start_timestamp: 1609459200000000000
        end_timestamp: 1609459200500000000
        attributes:
          http.method: "GET"
          http.url: "https://example.com/api/test"
        events:
          - timestamp: 1609459200250000000
            name: "exception"
            attributes:
              exception.type: "RuntimeError"
              exception.message: "Test error"
        links:
          - trace_id: "98765432109876543210987654321098"
            span_id: "8765432187654321"
            attributes:
              link.type: "child"
        status:
          code: 2
          message: "Internal error"
        flags: 1
`

	rt, err := ParseTraces([]byte(yamlData))
	assert.NoError(t, err)
	assert.NotNil(t, rt)

	// Verify resource
	assert.Equal(t, "my-service", rt.Resource["service.name"])
	assert.Equal(t, "1.0.0", rt.Resource["service.version"])

	// Verify scope traces
	assert.Len(t, rt.ScopeTraces, 1)
	scope := rt.ScopeTraces[0]
	assert.Equal(t, "my.tracer", scope.Name)
	assert.Equal(t, "1.0", scope.Version)
	assert.Equal(t, "https://opentelemetry.io/schemas/1.30.0", scope.SchemaURL)
	assert.Equal(t, "my-library", scope.Attributes["library.name"])

	// Verify spans
	assert.Len(t, scope.Spans, 1)
	span := scope.Spans[0]
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, "12345678901234567890123456789012", span.TraceID)
	assert.Equal(t, "1234567890123456", span.SpanID)
	assert.Equal(t, "9876543210987654", span.ParentSpanID)
	assert.Equal(t, int32(1), span.Kind)
	assert.Equal(t, int64(1609459200000000000), span.StartTimestamp)
	assert.Equal(t, int64(1609459200500000000), span.EndTimestamp)
	assert.Equal(t, "GET", span.Attributes["http.method"])
	assert.Equal(t, "https://example.com/api/test", span.Attributes["http.url"])

	// Verify events
	assert.Len(t, span.Events, 1)
	event := span.Events[0]
	assert.Equal(t, int64(1609459200250000000), event.Timestamp)
	assert.Equal(t, "exception", event.Name)
	assert.Equal(t, "RuntimeError", event.Attributes["exception.type"])
	assert.Equal(t, "Test error", event.Attributes["exception.message"])

	// Verify links
	assert.Len(t, span.Links, 1)
	link := span.Links[0]
	assert.Equal(t, "98765432109876543210987654321098", link.TraceID)
	assert.Equal(t, "8765432187654321", link.SpanID)
	assert.Equal(t, "child", link.Attributes["link.type"])

	// Verify status
	assert.Equal(t, int32(2), span.Status.Code)
	assert.Equal(t, "Internal error", span.Status.Message)

	// Verify flags
	assert.Equal(t, uint32(1), span.Flags)
}

func TestParseTracesWithJSON(t *testing.T) {
	jsonData := `{
		"resource": {
			"service.name": "json-service"
		},
		"scopes": [{
			"spans": [{
				"name": "JSON span",
				"trace_id": "12345678901234567890123456789012",
				"span_id": "1234567890123456"
			}]
		}]
	}`

	rt, err := ParseTraces([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "json-service", rt.Resource["service.name"])
	assert.Equal(t, "JSON span", rt.ScopeTraces[0].Spans[0].Name)
}

func TestMustParseTraces(t *testing.T) {
	// Test successful parsing
	yamlData := `
resource:
  service.name: "test"
scopes:
  - spans:
      - name: "test span"
`
	rt := MustParseTraces([]byte(yamlData))
	assert.NotNil(t, rt)
	assert.Equal(t, "test", rt.Resource["service.name"])

	// Test panic on invalid YAML
	assert.Panics(t, func() {
		MustParseTraces([]byte("invalid: yaml: content: ["))
	})
}

func TestTracesBuilder_AddFromYAML(t *testing.T) {
	builder := NewTracesBuilder()

	yamlData := `
resource:
  service.name: "test-service"
scopes:
  - spans:
      - name: "Test span"
        trace_id: "12345678901234567890123456789012"
        span_id: "1234567890123456"
        kind: 3
        start_timestamp: 1609459200000000000
        end_timestamp: 1609459200500000000
        attributes:
          http.method: "POST"
          http.status_code: 200
`

	err := builder.AddFromYAML([]byte(yamlData))
	assert.NoError(t, err)

	traces := builder.Build()
	assert.Equal(t, 1, traces.ResourceSpans().Len())

	resource := traces.ResourceSpans().At(0)
	serviceName, exists := resource.Resource().Attributes().Get(string(semconv.ServiceNameKey))
	assert.True(t, exists)
	assert.Equal(t, "test-service", serviceName.Str())

	assert.Equal(t, 1, resource.ScopeSpans().Len())
	scope := resource.ScopeSpans().At(0)
	assert.Equal(t, 1, scope.Spans().Len())

	span := scope.Spans().At(0)
	assert.Equal(t, "Test span", span.Name())
	assert.Equal(t, ptrace.SpanKind(3), span.Kind())
	assert.Equal(t, pcommon.Timestamp(1609459200000000000), span.StartTimestamp())
	assert.Equal(t, pcommon.Timestamp(1609459200500000000), span.EndTimestamp())

	method, exists := span.Attributes().Get("http.method")
	assert.True(t, exists)
	assert.Equal(t, "POST", method.Str())

	statusCode, exists := span.Attributes().Get("http.status_code")
	assert.True(t, exists)
	assert.Equal(t, int64(200), statusCode.Int())

	// Verify trace and span IDs
	assert.NotEqual(t, pcommon.NewTraceIDEmpty(), span.TraceID())
	assert.NotEqual(t, pcommon.NewSpanIDEmpty(), span.SpanID())
}

func TestSpanAllFields(t *testing.T) {
	yamlData := `
resource:
  service.name: "test-service"
scopes:
  - name: "test.tracer"
    version: "1.0"
    spans:
      - name: "Complete span with all fields"
        trace_id: "12345678901234567890123456789012"
        span_id: "1234567890123456"
        parent_span_id: "9876543210987654"
        kind: 2
        start_timestamp: 1609459200000000000
        end_timestamp: 1609459200500000000
        attributes:
          http.method: "GET"
          http.status_code: 200
        dropped_attributes_count: 1
        events:
          - timestamp: 1609459200100000000
            name: "start"
            attributes:
              event.id: "1"
          - timestamp: 1609459200400000000
            name: "end"
            attributes:
              event.id: "2"
            dropped_attributes_count: 1
        dropped_events_count: 1
        links:
          - trace_id: "98765432109876543210987654321098"
            span_id: "8765432187654321"
            trace_state: "vendor=test"
            attributes:
              link.relation: "follows"
            dropped_attributes_count: 1
        dropped_links_count: 1
        status:
          code: 1
          message: "Success"
        flags: 1
`

	builder := NewTracesBuilder()
	err := builder.AddFromYAML([]byte(yamlData))
	assert.NoError(t, err)

	traces := builder.Build()
	resource := traces.ResourceSpans().At(0)
	scope := resource.ScopeSpans().At(0)
	span := scope.Spans().At(0)

	// Test all basic fields
	assert.Equal(t, "Complete span with all fields", span.Name())
	assert.Equal(t, ptrace.SpanKind(2), span.Kind())
	assert.Equal(t, pcommon.Timestamp(1609459200000000000), span.StartTimestamp())
	assert.Equal(t, pcommon.Timestamp(1609459200500000000), span.EndTimestamp())

	// Test attributes
	method, exists := span.Attributes().Get("http.method")
	assert.True(t, exists)
	assert.Equal(t, "GET", method.Str())

	statusCode, exists := span.Attributes().Get("http.status_code")
	assert.True(t, exists)
	assert.Equal(t, int64(200), statusCode.Int())

	// Test dropped attributes count
	assert.Equal(t, uint32(1), span.DroppedAttributesCount())

	// Test events
	assert.Equal(t, 2, span.Events().Len())
	event1 := span.Events().At(0)
	assert.Equal(t, pcommon.Timestamp(1609459200100000000), event1.Timestamp())
	assert.Equal(t, "start", event1.Name())

	eventId, exists := event1.Attributes().Get("event.id")
	assert.True(t, exists)
	assert.Equal(t, "1", eventId.Str())

	event2 := span.Events().At(1)
	assert.Equal(t, uint32(1), event2.DroppedAttributesCount())

	// Test dropped events count
	assert.Equal(t, uint32(1), span.DroppedEventsCount())

	// Test links
	assert.Equal(t, 1, span.Links().Len())
	link := span.Links().At(0)
	assert.NotEqual(t, pcommon.NewTraceIDEmpty(), link.TraceID())
	assert.NotEqual(t, pcommon.NewSpanIDEmpty(), link.SpanID())

	linkRelation, exists := link.Attributes().Get("link.relation")
	assert.True(t, exists)
	assert.Equal(t, "follows", linkRelation.Str())

	assert.Equal(t, uint32(1), link.DroppedAttributesCount())

	// Test dropped links count
	assert.Equal(t, uint32(1), span.DroppedLinksCount())

	// Test status
	status := span.Status()
	assert.Equal(t, ptrace.StatusCode(1), status.Code())
	assert.Equal(t, "Success", status.Message())

	// Test flags
	assert.Equal(t, uint32(1), span.Flags())

	// Test trace/span IDs
	assert.NotEqual(t, pcommon.NewTraceIDEmpty(), span.TraceID())
	assert.NotEqual(t, pcommon.NewSpanIDEmpty(), span.SpanID())
	assert.NotEqual(t, pcommon.NewSpanIDEmpty(), span.ParentSpanID())
}

func TestInvalidTraceSpanIDsInTraces(t *testing.T) {
	builder := NewTracesBuilder()

	// Test invalid trace ID
	rt := &ResourceTraces{
		ScopeTraces: []ScopeTraces{
			{
				Spans: []Span{
					{
						Name:    "Test span",
						TraceID: "invalid_hex",
					},
				},
			},
		},
	}
	err := builder.Add(rt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid trace_id")

	// Test invalid span ID
	rt2 := &ResourceTraces{
		ScopeTraces: []ScopeTraces{
			{
				Spans: []Span{
					{
						Name:   "Test span",
						SpanID: "invalid_hex",
					},
				},
			},
		},
	}
	err = builder.Add(rt2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid span_id")

	// Test valid IDs
	validTraceID := hex.EncodeToString([]byte("1234567890123456"))
	validSpanID := hex.EncodeToString([]byte("12345678"))
	rt3 := &ResourceTraces{
		ScopeTraces: []ScopeTraces{
			{
				Spans: []Span{
					{
						Name:    "Test span",
						TraceID: validTraceID,
						SpanID:  validSpanID,
					},
				},
			},
		},
	}
	err = builder.Add(rt3)
	assert.NoError(t, err)
}
