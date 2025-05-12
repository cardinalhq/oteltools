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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestTracesBuilder_Resource(t *testing.T) {
	// Create a new TracesBuilder
	tb := NewTracesBuilder()

	// Create a pcommon.Map with some attributes
	attrs := pcommon.NewMap()
	attrs.PutStr("key1", "value1")
	attrs.PutInt("key2", 42)

	// Call the Resource method
	resourceBuilder := tb.Resource(attrs)

	// Verify that the returned TraceResourceBuilder is not nil
	assert.NotNil(t, resourceBuilder)

	// Verify that the resource attributes were copied correctly
	resourceSpans := tb.pm.ResourceSpans()
	assert.Equal(t, 1, resourceSpans.Len())

	resource := resourceSpans.At(0).Resource()
	copiedAttrs := resource.Attributes()
	assert.Equal(t, 2, copiedAttrs.Len())
	v, ok := copiedAttrs.Get("key1")
	require.True(t, ok)
	assert.Equal(t, "value1", v.Str())
	v, ok = copiedAttrs.Get("key2")
	require.True(t, ok)
	assert.Equal(t, int64(42), v.Int())

	// Call Resource again with the same attributes and verify it returns the same builder
	sameResourceBuilder := tb.Resource(attrs)
	assert.Equal(t, resourceBuilder, sameResourceBuilder)

	// Verify that no new ResourceSpans were added
	assert.Equal(t, 1, tb.pm.ResourceSpans().Len())
}

func TestRemoveEmptyTraces(t *testing.T) {
	// Create a new ptrace.Traces instance
	traces := ptrace.NewTraces()

	// Add a ResourceSpans with no ScopeSpans
	_ = traces.ResourceSpans().AppendEmpty()

	// Add a ResourceSpans with an empty ScopeSpans
	rs2 := traces.ResourceSpans().AppendEmpty()
	rs2.ScopeSpans().AppendEmpty()

	// Add a ResourceSpans with a ScopeSpans containing a Span
	rs3 := traces.ResourceSpans().AppendEmpty()
	scopeSpans := rs3.ScopeSpans().AppendEmpty()
	span := scopeSpans.Spans().AppendEmpty()
	span.SetName("test-span")

	// Call removeEmptyTraces
	result := removeEmptyTraces(traces)

	// Verify the result
	assert.Equal(t, 1, result.ResourceSpans().Len())

	// Verify that the remaining ResourceSpans is the one with the non-empty ScopeSpans
	remainingRS := result.ResourceSpans().At(0)
	assert.Equal(t, 1, remainingRS.ScopeSpans().Len())
	remainingScopeSpans := remainingRS.ScopeSpans().At(0)
	assert.Equal(t, 1, remainingScopeSpans.Spans().Len())
	assert.Equal(t, "test-span", remainingScopeSpans.Spans().At(0).Name())
}
