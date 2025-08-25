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

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func TestParse(t *testing.T) {
	yamlData := `
resource:
  service.name: "my-service"
  service.version: "1.0.0"
scopes:
  - name: "my.logger"
    version: "1.0"
    schema_url: "https://opentelemetry.io/schemas/1.30.0"
    attributes:
      library.name: "my-library"
    records:
      - timestamp: 1609459200000000000
        severity_text: "INFO"
        severity_number: 9
        body: "This is a test log message"
        attributes:
          user.id: "12345"
          request.method: "GET"
        trace_id: "12345678901234567890123456789012"
        span_id: "1234567890123456"
`

	rl, err := Parse([]byte(yamlData))
	assert.NoError(t, err)
	assert.NotNil(t, rl)

	// Verify resource
	assert.Equal(t, "my-service", rl.Resource["service.name"])
	assert.Equal(t, "1.0.0", rl.Resource["service.version"])

	// Verify scope logs
	assert.Len(t, rl.ScopeLogs, 1)
	scope := rl.ScopeLogs[0]
	assert.Equal(t, "my.logger", scope.Name)
	assert.Equal(t, "1.0", scope.Version)
	assert.Equal(t, "https://opentelemetry.io/schemas/1.30.0", scope.SchemaURL)
	assert.Equal(t, "my-library", scope.Attributes["library.name"])

	// Verify log records
	assert.Len(t, scope.LogRecords, 1)
	record := scope.LogRecords[0]
	assert.Equal(t, int64(1609459200000000000), record.Timestamp)
	assert.Equal(t, "INFO", record.SeverityText)
	assert.Equal(t, int32(9), record.SeverityNumber)
	assert.Equal(t, "This is a test log message", record.Body)
	assert.Equal(t, "12345", record.Attributes["user.id"])
	assert.Equal(t, "GET", record.Attributes["request.method"])
	assert.Equal(t, "12345678901234567890123456789012", record.TraceID)
	assert.Equal(t, "1234567890123456", record.SpanID)
}

func TestParseWithJSON(t *testing.T) {
	jsonData := `{
		"resource": {
			"service.name": "json-service"
		},
		"scopes": [{
			"records": [{
				"body": "JSON log message",
				"severity_text": "ERROR"
			}]
		}]
	}`

	rl, err := Parse([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "json-service", rl.Resource["service.name"])
	assert.Equal(t, "JSON log message", rl.ScopeLogs[0].LogRecords[0].Body)
	assert.Equal(t, "ERROR", rl.ScopeLogs[0].LogRecords[0].SeverityText)
}

func TestMustParse(t *testing.T) {
	// Test successful parsing
	yamlData := `
resource:
  service.name: "test"
scopes:
  - records:
      - body: "message"
`
	rl := MustParse([]byte(yamlData))
	assert.NotNil(t, rl)
	assert.Equal(t, "test", rl.Resource["service.name"])

	// Test panic on invalid YAML
	assert.Panics(t, func() {
		MustParse([]byte("invalid: yaml: content: ["))
	})
}

func TestFromRaw(t *testing.T) {
	attrs := map[string]any{
		"string_val":  "test_value",
		"bool_val":    true,
		"int_val":     42,
		"float_val":   3.14,
		"nested_val":  map[string]any{"key": "value"},
		"array_val":   []any{"a", 1, true},
		"bytes_val":   []byte("hello"),
	}

	result, err := fromRaw(attrs)
	assert.NoError(t, err)
	assert.Equal(t, 7, result.Len())

	// Check converted values
	stringVal, exists := result.Get("string_val")
	assert.True(t, exists)
	assert.Equal(t, "test_value", stringVal.Str())

	boolVal, exists := result.Get("bool_val")
	assert.True(t, exists)
	assert.Equal(t, true, boolVal.Bool())

	intVal, exists := result.Get("int_val")
	assert.True(t, exists)
	assert.Equal(t, int64(42), intVal.Int())

	floatVal, exists := result.Get("float_val")
	assert.True(t, exists)
	assert.InDelta(t, 3.14, floatVal.Double(), 0.001)

	nestedVal, exists := result.Get("nested_val")
	assert.True(t, exists)
	nestedMap := nestedVal.Map()
	keyVal, keyExists := nestedMap.Get("key")
	assert.True(t, keyExists)
	assert.Equal(t, "value", keyVal.Str())

	arrayVal, exists := result.Get("array_val")
	assert.True(t, exists)
	arraySlice := arrayVal.Slice()
	assert.Equal(t, 3, arraySlice.Len())
	assert.Equal(t, "a", arraySlice.At(0).Str())
	assert.Equal(t, int64(1), arraySlice.At(1).Int())
	assert.Equal(t, true, arraySlice.At(2).Bool())

	bytesVal, exists := result.Get("bytes_val")
	assert.True(t, exists)
	assert.Equal(t, []byte("hello"), bytesVal.Bytes().AsRaw())
}

func TestFromRawCapabilities(t *testing.T) {
	// Test what FromRaw can and cannot handle
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{"string", "test", false},
		{"int", 42, false},
		{"float", 3.14, false},
		{"bool", true, false},
		{"map", map[string]any{"key": "value"}, false},
		{"slice", []any{"a", "b"}, false},
		{"bytes", []byte("hello"), false},
		{"nil", nil, false},
		{"complex", complex(1, 2), true}, // Should fail
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := pcommon.NewValueEmpty()
			err := val.FromRaw(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "Expected FromRaw to fail for %T", tt.input)
			} else {
				assert.NoError(t, err, "Expected FromRaw to succeed for %T", tt.input)
			}
		})
	}
}

func TestLogBuilder_AddFromYAML(t *testing.T) {
	builder := NewLogBuilder()
	
	yamlData := `
resource:
  service.name: "test-service"
scopes:
  - records:
      - body: "Test message"
        severity_text: "INFO"
        attributes:
          key: "value"
`

	err := builder.AddFromYAML([]byte(yamlData))
	assert.NoError(t, err)

	logs := builder.Build()
	assert.Equal(t, 1, logs.ResourceLogs().Len())
	
	resource := logs.ResourceLogs().At(0)
	serviceName, exists := resource.Resource().Attributes().Get(string(semconv.ServiceNameKey))
	assert.True(t, exists)
	assert.Equal(t, "test-service", serviceName.Str())
	
	assert.Equal(t, 1, resource.ScopeLogs().Len())
	scope := resource.ScopeLogs().At(0)
	assert.Equal(t, 1, scope.LogRecords().Len())
	
	record := scope.LogRecords().At(0)
	assert.Equal(t, "Test message", record.Body().Str())
	assert.Equal(t, "INFO", record.SeverityText())
	
	attrVal, exists := record.Attributes().Get("key")
	assert.True(t, exists)
	assert.Equal(t, "value", attrVal.Str())
}

func TestInvalidTraceSpanIDs(t *testing.T) {
	builder := NewLogBuilder()
	// Test invalid trace ID
	rl := &ResourceLogs{
		ScopeLogs: []ScopeLogs{
			{
				LogRecords: []LogRecord{
					{
						Body:    "Test message",
						TraceID: "invalid_hex",
					},
				},
			},
		},
	}
	err := builder.Add(rl)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid trace_id")

	// Test invalid span ID  
	rl2 := &ResourceLogs{
		ScopeLogs: []ScopeLogs{
			{
				LogRecords: []LogRecord{
					{
						Body:   "Test message",
						SpanID: "invalid_hex",
					},
				},
			},
		},
	}
	err = builder.Add(rl2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid span_id")

	// Test wrong length trace ID
	rl3 := &ResourceLogs{
		ScopeLogs: []ScopeLogs{
			{
				LogRecords: []LogRecord{
					{
						Body:    "Test message",  
						TraceID: "1234", // Too short
					},
				},
			},
		},
	}
	err = builder.Add(rl3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 32 hex characters")

	// Test wrong length span ID
	rl4 := &ResourceLogs{
		ScopeLogs: []ScopeLogs{
			{
				LogRecords: []LogRecord{
					{
						Body:   "Test message",
						SpanID: "1234", // Too short
					},
				},
			},
		},
	}
	err = builder.Add(rl4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 16 hex characters")

	// Test valid IDs
	validTraceID := hex.EncodeToString([]byte("1234567890123456"))
	validSpanID := hex.EncodeToString([]byte("12345678"))
	rl5 := &ResourceLogs{
		ScopeLogs: []ScopeLogs{
			{
				LogRecords: []LogRecord{
					{
						Body:    "Test message",
						TraceID: validTraceID,
						SpanID:  validSpanID,
					},
				},
			},
		},
	}
	err = builder.Add(rl5)
	assert.NoError(t, err)
}

func TestLogRecordAllFields(t *testing.T) {
	yamlData := `
resource:
  service.name: "test-service"
scopes:
  - name: "test.logger"
    version: "1.0"
    records:
      - timestamp: 1609459200000000000
        observed_timestamp: 1609459200500000000
        severity_text: "ERROR"
        severity_number: 17
        body: "Complete log record with all fields"
        attributes:
          user.id: "12345"
          error.code: "E001"
        trace_id: "12345678901234567890123456789012"
        span_id: "1234567890123456"
        flags: 1
        dropped_attributes_count: 2
        event_name: "exception"
`

	builder := NewLogBuilder()
	err := builder.AddFromYAML([]byte(yamlData))
	assert.NoError(t, err)

	logs := builder.Build()
	assert.Equal(t, 1, logs.ResourceLogs().Len())
	
	resource := logs.ResourceLogs().At(0)
	assert.Equal(t, 1, resource.ScopeLogs().Len())
	
	scope := resource.ScopeLogs().At(0)
	assert.Equal(t, "test.logger", scope.Scope().Name())
	assert.Equal(t, "1.0", scope.Scope().Version())
	assert.Equal(t, 1, scope.LogRecords().Len())
	
	record := scope.LogRecords().At(0)
	
	// Test all timestamps
	assert.Equal(t, pcommon.Timestamp(1609459200000000000), record.Timestamp())
	assert.Equal(t, pcommon.Timestamp(1609459200500000000), record.ObservedTimestamp())
	
	// Test severity
	assert.Equal(t, "ERROR", record.SeverityText())
	assert.Equal(t, plog.SeverityNumber(17), record.SeverityNumber())
	
	// Test body
	assert.Equal(t, "Complete log record with all fields", record.Body().Str())
	
	// Test attributes
	userID, exists := record.Attributes().Get("user.id")
	assert.True(t, exists)
	assert.Equal(t, "12345", userID.Str())
	
	errorCode, exists := record.Attributes().Get("error.code")
	assert.True(t, exists)
	assert.Equal(t, "E001", errorCode.Str())
	
	// Test trace/span IDs
	assert.NotEqual(t, pcommon.NewTraceIDEmpty(), record.TraceID())
	assert.NotEqual(t, pcommon.NewSpanIDEmpty(), record.SpanID())
	
	// Test new fields
	assert.Equal(t, plog.LogRecordFlags(1), record.Flags())
	assert.Equal(t, uint32(2), record.DroppedAttributesCount())
	assert.Equal(t, "exception", record.EventName())
}