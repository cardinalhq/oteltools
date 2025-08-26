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
	"github.com/cardinalhq/oteltools/hashutils"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func attrkey(attr pcommon.Map) uint64 {
	if attr.Len() == 0 {
		return 1
	}
	m := attr.AsRaw()
	return hashutils.HashAny(nil, m)
}

func metrickey(name string, units string, ty pmetric.MetricType) uint64 {
	return hashutils.HashStrings(nil, name, units, ty.String())
}

func scopekey(name, version, schemaURL string, attr pcommon.Map) uint64 {
	if attr.Len() == 0 && name == "" && version == "" && schemaURL == "" {
		return 1
	}
	m := attr.AsRaw()
	// Include name, version, and schema URL in the hash for uniqueness
	return hashutils.HashAny(nil, map[string]any{
		"name":       name,
		"version":    version,
		"schema_url": schemaURL,
		"attributes": m,
	})
}
