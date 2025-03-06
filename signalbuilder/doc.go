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

// SignalBuilder is a package for building signals for OpenTelemetry
// signal types.  It maintains an internal fast lookup method so that
// data duplication is minimized, and the result will have exactly
// one Resource level object based on attributes, and each
// Resource will have exactly one of each Scope level object,
// and each Scope will have exactly one of the appropriate signal
// types.
//
// Sums are always configured to be non-monotonic and delta aggregation.
//
// Currently, the only signal type supported is Metric.
//
// This package is not thread-safe.
package signalbuilder
