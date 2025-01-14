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

package translate

const (
	CardinalFieldPrefix    = "_cardinalhq"
	CardinalFieldPrefixDot = CardinalFieldPrefix + "."

	CardinalFieldAggregate         = CardinalFieldPrefixDot + "aggregate"
	CardinalFieldBucketBounds      = CardinalFieldPrefixDot + "bucket_bounds"
	CardinalFieldClassification    = CardinalFieldPrefixDot + "classification"
	CardinalFieldCollectorID       = CardinalFieldPrefixDot + "collector_id"
	CardinalFieldCounts            = CardinalFieldPrefixDot + "counts"
	CardinalFieldCustomerID        = CardinalFieldPrefixDot + "customer_id"
	CardinalFieldCustomerName      = CardinalFieldPrefixDot + "customer_name"
	CardinalFieldDecoratorPodName  = CardinalFieldPrefixDot + "decorator.pod_name"
	CardinalFieldDropForVendor     = CardinalFieldPrefixDot + "drop_for_vendor"
	CardinalFieldDropMarker        = CardinalFieldPrefixDot + "drop_marker"
	CardinalFieldFiltered          = CardinalFieldPrefixDot + "filtered"
	CardinalFieldFilteredReason    = CardinalFieldPrefixDot + "filtered_reason"
	CardinalFieldFingerprint       = CardinalFieldPrefixDot + "fingerprint"
	CardinalFieldFingerprintError  = CardinalFieldPrefixDot + "fingerprint_error"
	CardinalFieldHostname          = CardinalFieldPrefixDot + "hostname"
	CardinalFieldID                = CardinalFieldPrefixDot + "id"
	CardinalFieldIsRootSpan        = CardinalFieldPrefixDot + "is_root_span"
	CardinalFieldJSON              = CardinalFieldPrefixDot + "json"
	CardinalFieldLevel             = CardinalFieldPrefixDot + "level"
	CardinalFieldMessage           = CardinalFieldPrefixDot + "message"
	CardinalFieldMetricType        = CardinalFieldPrefixDot + "metric_type"
	CardinalFieldName              = CardinalFieldPrefixDot + "name"
	CardinalFieldNegativeCounts    = CardinalFieldPrefixDot + "negative_counts"
	CardinalFieldPositiveCounts    = CardinalFieldPrefixDot + "positive_counts"
	CardinalFieldReceiverType      = CardinalFieldPrefixDot + "receiver_type"
	CardinalFieldResourceName      = CardinalFieldPrefixDot + "resource_name"
	CardinalFieldResourceSchemaURL = CardinalFieldPrefixDot + "resource_schemaurl"
	CardinalFieldRuleConfig        = CardinalFieldPrefixDot + "ruleconfig"
	CardinalFieldRuleID            = CardinalFieldPrefixDot + "rule_id"
	CardinalFieldRuleMatch         = CardinalFieldPrefixDot + "rule_match"
	CardinalFieldRulesMatched      = CardinalFieldPrefixDot + "rules_matched"
	CardinalFieldScale             = CardinalFieldPrefixDot + "scale"
	CardinalFieldScopeSchemaURL    = CardinalFieldPrefixDot + "scope_schemaurl"
	CardinalFieldSpanDuration      = CardinalFieldPrefixDot + "span_duration"
	CardinalFieldSpanEndTime       = CardinalFieldPrefixDot + "span_end_time"
	CardinalFieldSpanEventcount    = CardinalFieldPrefixDot + "span_eventcount"
	CardinalFieldSpanHasError      = CardinalFieldPrefixDot + "span_has_error"
	CardinalFieldSpanIsSlow        = CardinalFieldPrefixDot + "isSlow"
	CardinalFieldSpanKind          = CardinalFieldPrefixDot + "span_kind"
	CardinalFieldSpanName          = CardinalFieldPrefixDot + "span_name"
	CardinalFieldSpanParentSpanID  = CardinalFieldPrefixDot + "span_parent_span_id"
	CardinalFieldSpanSpanID        = CardinalFieldPrefixDot + "span_span_id"
	CardinalFieldSpanStartTime     = CardinalFieldPrefixDot + "span_start_time"
	CardinalFieldSpanStatusCode    = CardinalFieldPrefixDot + "span_status_code"
	CardinalFieldSpanStatusMessage = CardinalFieldPrefixDot + "span_status_message"
	CardinalFieldSpanTraceID       = CardinalFieldPrefixDot + "span_trace_id"
	CardinalFieldTelemetryType     = CardinalFieldPrefixDot + "telemetry_type"
	CardinalFieldTID               = CardinalFieldPrefixDot + "tid"
	CardinalFieldTimestamp         = CardinalFieldPrefixDot + "timestamp"
	CardinalFieldTokenMap          = CardinalFieldPrefixDot + "tokenMap"
	CardinalFieldTokens            = CardinalFieldPrefixDot + "tokens"
	CardinalFieldTraceHasError     = CardinalFieldPrefixDot + "trace_has_error"
	CardinalFieldValue             = CardinalFieldPrefixDot + "value"
	CardinalFieldWouldFilter       = CardinalFieldPrefixDot + "would_filter"
	CardinalFieldZeroCount         = CardinalFieldPrefixDot + "zero_count"
	ConditionsMatched              = "conditions_matched"
	ErrorMsg                       = "errorMsg"
	RuleId                         = "rule_id"
	SamplerAllowed                 = "sampler_allowed"
	Stage                          = "stage"
	StatementsEvaluated            = "statements_evaluated"
	Version                        = "version"

	CardinalMetricTypeCount                = "count"
	CardinalMetricTypeExponentialHistogram = "exponential_histogram"
	CardinalMetricTypeGauge                = "gauge"
	CardinalMetricTypeHistogram            = "histogram"

	CardinalTelemetryTypeLogs    = "logs"
	CardinalTelemetryTypeMetrics = "metrics"
	CardinalTelemetryTypeTraces  = "traces"

	CardinalHeaderAPIKey   = "X-CardinalHQ-API-Key"
	CardinalHeaderDDAPIKey = "DD-API-KEY"
)
