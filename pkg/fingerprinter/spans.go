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

package fingerprinter

import (
	"github.com/cardinalhq/oteltools/pkg/ottl/functions"
	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"strconv"
	"strings"
)

func CalculateSpanFingerprint(sr ptrace.Span, fpr Fingerprinter) int64 {
	fingerprintAttributes := make([]string, 0)

	var exceptionMessage string
	for i := 0; i < sr.Events().Len(); i++ {
		event := sr.Events().At(i)
		if event.Name() == semconv.ExceptionEventName {
			if exType, found := event.Attributes().Get(string(semconv.ExceptionTypeKey)); found {
				exceptionMessage = exceptionMessage + exType.AsString()
			}
			if exMsg, found := event.Attributes().Get(string(semconv.ExceptionMessageKey)); found {
				if exceptionMessage != "" {
					exceptionMessage = exceptionMessage + " " + exMsg.AsString()
				} else {
					exceptionMessage = exMsg.AsString()
				}
			}
			if exStack, found := event.Attributes().Get(string(semconv.ExceptionStacktraceKey)); found {
				if exceptionMessage != "" {
					exceptionMessage = exceptionMessage + "\n" + exStack.AsString()
				} else {
					exceptionMessage = exStack.AsString()
				}
			}
			break
		}
	}
	if exceptionMessage != "" {
		computedFp, _, _, _, err := fpr.Fingerprint(exceptionMessage)
		if err == nil {
			fingerprintAttributes = append(fingerprintAttributes, strconv.FormatInt(computedFp, 10))
		}
	}

	sanitizedName := functions.ScrubWord(sr.Name())
	fingerprintAttributes = append(fingerprintAttributes, sanitizedName)

	spanKindStr := sr.Kind().String()
	fingerprintAttributes = append(fingerprintAttributes, spanKindStr)

	statusCodeStr := sr.Status().Code().String()
	fingerprintAttributes = append(fingerprintAttributes, statusCodeStr)
	return int64(xxhash.Sum64String(strings.Join(fingerprintAttributes, "##")))

}
