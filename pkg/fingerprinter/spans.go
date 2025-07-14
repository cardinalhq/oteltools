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
	"github.com/cardinalhq/oteltools/hashutils"
	"github.com/cardinalhq/oteltools/pkg/translate"
	"github.com/cespare/xxhash/v2"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"strings"
)

const (
	serviceNameKey   = string(semconv.ServiceNameKey)
	clusterNameKey   = string(semconv.K8SClusterNameKey)
	namespaceNameKey = string(semconv.K8SNamespaceNameKey)
)

func ComputeExemplarKey(rl pcommon.Resource, extraKeys []string) ([]string, int64) {
	keys := []string{
		clusterNameKey, GetFromResource(rl.Attributes(), clusterNameKey),
		namespaceNameKey, GetFromResource(rl.Attributes(), namespaceNameKey),
		serviceNameKey, GetFromResource(rl.Attributes(), serviceNameKey),
	}
	keys = append(keys, extraKeys...)
	return keys, int64(hashutils.HashStrings(nil, keys...))
}

func GetFromResource(attr pcommon.Map, key string) string {
	clusterVal, clusterFound := attr.Get(key)
	if !clusterFound {
		return "unknown"
	}
	return clusterVal.AsString()
}

func GetFingerprintAttribute(l pcommon.Map) int64 {
	fnk := translate.CardinalFieldFingerprint
	if fingerprintField, found := l.Get(fnk); found {
		return fingerprintField.Int()
	}
	return 0
}

func GetExceptionMessage(sr ptrace.Span) string {
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
	return exceptionMessage
}

func GetStringAttribute(sr ptrace.Span, key string) string {
	attrValue, found := sr.Attributes().Get(key)
	if !found {
		return ""
	}
	return attrValue.AsString()
}

func CalculateSpanFingerprint(res pcommon.Resource, sr ptrace.Span) int64 {
	fingerprintAttributes := make([]string, 0)
	clusterName := GetFromResource(res.Attributes(), clusterNameKey)
	namespaceName := GetFromResource(res.Attributes(), namespaceNameKey)
	serviceName := GetFromResource(res.Attributes(), serviceNameKey)
	spanKindStr := sr.Kind().String()
	fingerprintAttributes = append(fingerprintAttributes, clusterName, namespaceName, serviceName, spanKindStr)

	dbSystem := GetStringAttribute(sr, string(semconv.DBSystemNameKey))
	messagingSystem := GetStringAttribute(sr, string(semconv.MessagingSystemKey))
	httpRequestMethod := GetStringAttribute(sr, string(semconv.HTTPRequestMethodKey))

	if messagingSystem != "" {
		messagingOperationType := GetStringAttribute(sr, string(semconv.MessagingOperationTypeKey))
		messagingDestinationName := GetStringAttribute(sr, string(semconv.MessagingDestinationNameKey))
		fingerprintAttributes = append(fingerprintAttributes, messagingSystem, messagingOperationType, messagingDestinationName)
		return toHash(fingerprintAttributes)
	}
	if dbSystem != "" {
		dbNamespace := GetStringAttribute(sr, string(semconv.DBNamespaceKey))
		dbOperationName := GetStringAttribute(sr, string(semconv.DBOperationNameKey))
		serverAddress := GetStringAttribute(sr, string(semconv.ServerAddressKey))
		collectionName := GetStringAttribute(sr, string(semconv.DBCollectionNameKey))
		fingerprintAttributes = append(fingerprintAttributes, sr.Name(), dbSystem, dbNamespace, dbOperationName, serverAddress, collectionName)
		return toHash(fingerprintAttributes)
	}
	if httpRequestMethod != "" {
		httpUrlTemplate := GetStringAttribute(sr, string(semconv.URLTemplateKey))
		fingerprintAttributes = append(fingerprintAttributes, httpRequestMethod, httpUrlTemplate)
		return toHash(fingerprintAttributes)
	}

	sanitizedName := sr.Name()
	fingerprintAttributes = append(fingerprintAttributes, sanitizedName)

	return toHash(fingerprintAttributes)
}

func toHash(fingerprintAttributes []string) int64 {
	return int64(xxhash.Sum64String(strings.Join(fingerprintAttributes, "##")))
}
