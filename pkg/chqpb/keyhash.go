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

package chqpb

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cespare/xxhash"
)

func (key *LogIngestStat_Key) Hash() uint64 {
	sb := strings.Builder{}
	sb.WriteString(key.Service)
	sb.WriteRune(':')
	sb.WriteString(strconv.FormatInt(key.Fingerprint, 10))
	sb.WriteRune(':')
	sb.WriteString(strconv.FormatInt(key.TsHour, 10))
	sb.WriteRune(':')
	sb.WriteString(key.CustomerId)
	sb.WriteRune(':')
	sb.WriteString(key.ProcessorId)
	sb.WriteRune(':')
	sb.WriteString(key.PhaseType)
	sb.WriteRune(':')
	sb.WriteString(key.CollectorId)

	appendAttributes(&sb, key.Attributes)

	return xxhash.Sum64String(sb.String())
}

func (key *MetricIngestStat_Key) Hash() uint64 {
	sb := strings.Builder{}
	sb.WriteString(key.Service)
	sb.WriteRune(':')
	sb.WriteString(key.MetricName)
	sb.WriteRune(':')
	sb.WriteString(key.TagName)
	sb.WriteRune(':')
	sb.WriteString(strconv.FormatInt(key.TsHour, 10))
	sb.WriteRune(':')
	sb.WriteString(key.CustomerId)
	sb.WriteRune(':')
	sb.WriteString(key.ProcessorId)
	sb.WriteRune(':')
	sb.WriteString(key.PhaseType)
	sb.WriteRune(':')
	sb.WriteString(key.CollectorId)
	sb.WriteRune(':')
	sb.WriteString(key.TagScope)
	sb.WriteRune(':')
	sb.WriteString(key.MetricType)

	appendAttributes(&sb, key.Attributes)

	return xxhash.Sum64String(sb.String())
}

func (key *SpanIngestStat_Key) Hash() uint64 {
	sb := strings.Builder{}
	sb.WriteString(key.Service)
	sb.WriteRune(':')
	sb.WriteString(strconv.FormatInt(key.Fingerprint, 10))
	sb.WriteRune(':')
	sb.WriteString(strconv.FormatInt(key.TsHour, 10))
	sb.WriteRune(':')
	sb.WriteString(key.CustomerId)
	sb.WriteRune(':')
	sb.WriteString(key.ProcessorId)
	sb.WriteRune(':')
	sb.WriteString(key.PhaseType)
	sb.WriteRune(':')
	sb.WriteString(key.CollectorId)

	attributes := key.Attributes
	appendAttributes(&sb, attributes)

	return xxhash.Sum64String(sb.String())
}

func appendAttributes(sb *strings.Builder, attrs []*Attribute) {
	if len(attrs) == 0 {
		return
	}
	ret := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		ret = append(ret, AttributeToString(attr))
	}
	slices.Sort(ret)
	for _, a := range ret {
		sb.WriteString(a)
	}
}

func AttributeToString(attr *Attribute) string {
	return fmt.Sprintf(":%s=%s=%d=%t", attr.Key, attr.Value, attr.Type, attr.IsAttribute)
}
