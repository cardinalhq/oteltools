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

package telemetry

import "go.opentelemetry.io/otel/metric"

type DeferrableHistogram[T int64] interface {
	record(delta T, options ...metric.RecordOption)
}

func HistogramRecord[T int64](histogram DeferrableHistogram[T], delta T, options ...metric.RecordOption) {
	if histogram == nil {
		return
	}
	histogram.record(delta, options...)
}
