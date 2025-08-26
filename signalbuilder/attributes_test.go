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
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestFromRaw(t *testing.T) {
	attrs := map[string]any{
		"string_val": "test_value",
		"bool_val":   true,
		"int_val":    42,
		"float_val":  3.14,
		"nested_val": map[string]any{"key": "value"},
		"array_val":  []any{"a", 1, true},
		"bytes_val":  []byte("hello"),
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

func TestFromRawNilAndEmptyMaps(t *testing.T) {
	// Test nil map
	result, err := fromRaw(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Len())

	// Test empty map
	result, err = fromRaw(map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, 0, result.Len())
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
