// Copyright 2024 CardinalHQ, Inc
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
	"strings"
	"testing"

	"github.com/cespare/xxhash"
)

func TestAppendTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []*Attribute
		expected string
	}{
		{
			name:     "empty tags",
			tags:     []*Attribute{},
			expected: "",
		},
		{
			name: "single tag",
			tags: []*Attribute{
				{
					Key:   "key1",
					Value: "value1",
				},
			},
			expected: `:key1=value1=0=false`,
		},
		{
			name: "multiple tags",
			tags: []*Attribute{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "value2",
				},
			},
			expected: `:key1=value1=0=false:key2=value2=0=false`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			appendAttributes(&sb, tt.tags)
			if sb.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, sb.String())
			}
		})
	}
}

func TestLogIngestStat_Key_Hash(t *testing.T) {
	tests := []*struct {
		name     string
		key      LogIngestStat_Key
		expected uint64
	}{
		{
			name: "empty fields",
			key: LogIngestStat_Key{
				Service:     "",
				Fingerprint: 0,
				TsHour:      0,
				CustomerId:  "",
				PhaseType:   "",
				ProcessorId: "",
				CollectorId: "",
				Attributes:  make([]*Attribute, 0),
			},
			expected: xxhash.Sum64String(":0:0::::"),
		},
		{
			name: "filled fields without tags",
			key: LogIngestStat_Key{
				Service:     "service1",
				Fingerprint: 98754,
				TsHour:      1234567890,
				CustomerId:  "customer1",
				PhaseType:   "phase1",
				ProcessorId: "vendor1",
				CollectorId: "collector1",
				Attributes:  make([]*Attribute, 0),
			},
			expected: xxhash.Sum64String("service1:98754:1234567890:customer1:vendor1:phase1:collector1"),
		},
		{
			name: "filled fields with tags",
			key: LogIngestStat_Key{
				Service:     "service1",
				Fingerprint: 98754,
				TsHour:      1234567890,
				CustomerId:  "customer1",
				PhaseType:   "phase1",
				ProcessorId: "vendor1",
				CollectorId: "collector1",
				Attributes: []*Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
			expected: 5705948733490865915,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.Hash(); got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestMetricIngestStat_Key_Hash(t *testing.T) {
	tests := []*struct {
		name     string
		key      MetricIngestStat_Key
		expected uint64
	}{
		{
			name: "empty fields",
			key: MetricIngestStat_Key{
				Service:     "",
				MetricName:  "",
				TagName:     "",
				TsHour:      0,
				CustomerId:  "",
				PhaseType:   "",
				ProcessorId: "",
				CollectorId: "",
				Attributes:  make([]*Attribute, 0),
			},
			expected: xxhash.Sum64String(":::0::::::"),
		},
		{
			name: "filled fields without tags",
			key: MetricIngestStat_Key{
				Service:     "service1",
				MetricName:  "metric1",
				TagName:     "tag1",
				TsHour:      1234567890,
				CustomerId:  "customer1",
				PhaseType:   "phase1",
				ProcessorId: "vendor1",
				CollectorId: "collector1",
				TagScope:    "tagscope1",
				Attributes:  make([]*Attribute, 0),
			},
			expected: xxhash.Sum64String("service1:metric1:tag1:1234567890:customer1:vendor1:phase1:collector1:tagscope1:"),
		},
		{
			name: "filled fields with tags",
			key: MetricIngestStat_Key{
				Service:     "service1",
				MetricName:  "metric1",
				TagName:     "tag1",
				TsHour:      1234567890,
				CustomerId:  "customer1",
				PhaseType:   "phase1",
				ProcessorId: "vendor1",
				CollectorId: "collector1",
				Attributes: []*Attribute{
					{
						Key:   "key1",
						Value: "value1",
					},
					{
						Key:   "key2",
						Value: "value2",
					},
				},
			},
			expected: 4501433055337193504,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.Hash(); got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}
