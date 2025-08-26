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

// Package signalbuilder provides efficient builders for constructing OpenTelemetry signals
// (metrics, traces, and logs) with automatic deduplication and structured organization.
//
// # Overview
//
// SignalBuilder optimizes the creation of OpenTelemetry pdata structures by:
//   - Deduplicating resources based on attributes using fast hash-based lookups
//   - Organizing data hierarchically: Resource → Scope → Signal
//   - Minimizing memory allocation through internal caching
//   - Providing both protobuf and structured (JSON/YAML) output formats
//
// # Architecture
//
// The package follows OpenTelemetry's data model hierarchy:
//   Resource (service.name, host.name, etc.)
//     ├── Scope (instrumentation library)
//         ├── Metrics/Traces/Logs
//
// Each builder maintains a hash map of resources to ensure exactly one Resource
// object per unique set of attributes, and exactly one Scope per resource.
//
// # Supported Signal Types
//
//   - Metrics: All OTEL metric types (Gauge, Sum, Histogram, ExponentialHistogram, Summary)
//   - Traces: Span data with full trace context support
//   - Logs: Structured log records with severity levels and attributes
//
// # Usage Examples
//
// Basic metrics usage:
//   builder := NewMetricsBuilder()
//   resourceAttrs := map[string]any{"service.name": "my-service"}
//   scopeAttrs := map[string]any{"version": "1.0.0"}
//   
//   // Add a gauge metric
//   builder.AddGauge("cpu_usage", "percent", "CPU usage percentage",
//       resourceAttrs, "otelcol", "0.1.0", "", scopeAttrs,
//       map[string]any{"cpu": "cpu0"}, 85.5, time.Now().UnixNano())
//
// Export to different formats:
//   pmetrics := builder.Build()                    // OpenTelemetry pdata
//   jsonBytes := builder.BuildStructuredJSON()     // JSON format
//   yamlBytes := builder.BuildStructuredYAML()     // YAML format
//
// # Configuration Notes
//
//   - Sum metrics default to non-monotonic delta aggregation
//   - Histogram metrics use explicit bounds
//   - ExponentialHistogram metrics use base-2 exponential buckets
//   - All timestamps should be in nanoseconds since Unix epoch
//
// # Thread Safety
//
// This package is NOT thread-safe. Use separate builder instances for
// concurrent operations or provide external synchronization.
//
// # Performance Characteristics
//
//   - O(1) resource lookup via hash maps
//   - Memory-efficient through resource/scope deduplication  
//   - Optimized for high-throughput telemetry data processing
package signalbuilder
