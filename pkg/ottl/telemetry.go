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

package ottl

import (
	"github.com/cardinalhq/oteltools/pkg/telemetry"
	"go.opentelemetry.io/otel/metric"
)

// Telemetry is a struct that holds all the telemetry metrics for the OTTL processor.
type Telemetry struct {
	// ConditionsErrorCounter is a counter for the number of times condition evaluation errored.
	ConditionsErrorCounter telemetry.DeferrableCounter

	// ConditionsEvaluatedCounter is a counter for the number of times conditions were evaluated.
	ConditionsEvaluatedCounter telemetry.DeferrableCounter

	// ConditionsEvaluatedHistogram is a histogram for the timing of conditions evaluated.
	ConditionsEvaluatedHistogram telemetry.DeferrableHistogram

	// RateLimitedCounter is a counter for the number of times conditions passed, but the
	// statements were not executed due to rate limiting.
	RateLimitedCounter telemetry.DeferrableCounter

	// StatementsErrorCounter is a counter for the number of times statement processing errored.
	StatementsErrorCounter telemetry.DeferrableCounter

	// StatementsExecutedHistogram is a histogram for the number of times statements were executed.
	StatementsExecutedHistogram telemetry.DeferrableHistogram

	// StatementsExecutedCounter is a counter for the timing of statements executed.
	StatementsExecutedCounter telemetry.DeferrableCounter
}

func NewTelemetry(meter metric.Meter) *Telemetry {
	ret := &Telemetry{}

	if cec, err := telemetry.NewDeferrableInt64Counter(meter,
		"ottl.conditions.error.count",
		[]metric.Int64CounterOption{
			metric.WithDescription("The number of times condition evaluation errored."),
			metric.WithUnit("1"),
		}, nil); err == nil {
		ret.ConditionsErrorCounter = cec
	}
	if cec, err := telemetry.NewDeferrableInt64Counter(meter,
		"ottl.conditions.evaluated.count",
		[]metric.Int64CounterOption{
			metric.WithDescription("The number of times conditions were evaluated."),
			metric.WithUnit("1"),
		}, nil); err == nil {
		ret.ConditionsEvaluatedCounter = cec
	}
	if ceh, err := telemetry.NewDeferrableHistogram(meter,
		"ottl.conditions.evaluated.time",
		[]metric.Int64HistogramOption{
			metric.WithDescription("The timing of conditions evaluated."),
			metric.WithUnit("ns"),
		}, nil); err == nil {
		ret.ConditionsEvaluatedHistogram = ceh
	}
	if rlc, err := telemetry.NewDeferrableInt64Counter(meter,
		"ottl.rate_limited.count",
		[]metric.Int64CounterOption{
			metric.WithDescription("The number of times conditions passed, but the statements were not executed due to rate limiting."),
			metric.WithUnit("1"),
		}, nil); err == nil {
		ret.RateLimitedCounter = rlc
	}
	if sec, err := telemetry.NewDeferrableInt64Counter(meter,
		"ottl.statements.error.count",
		[]metric.Int64CounterOption{
			metric.WithDescription("The number of times statement processing errored."),
			metric.WithUnit("1"),
		}, nil); err == nil {
		ret.StatementsErrorCounter = sec
	}
	if seh, err := telemetry.NewDeferrableHistogram(meter,
		"ottl.statements.executed.time",
		[]metric.Int64HistogramOption{
			metric.WithDescription("The timing of statements executed."),
			metric.WithUnit("ns"),
		}, nil); err == nil {
		ret.StatementsExecutedHistogram = seh
	}
	if sec, err := telemetry.NewDeferrableInt64Counter(meter,
		"ottl.statements.executed.count",
		[]metric.Int64CounterOption{
			metric.WithDescription("The number of times statements were executed."),
			metric.WithUnit("1"),
		}, nil); err == nil {
		ret.StatementsExecutedCounter = sec
	}

	return ret
}
