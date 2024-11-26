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
	"testing"

	"github.com/cardinalhq/oteltools/pkg/stats"
	"github.com/stretchr/testify/assert"
)

func TestLogStats_Key(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		logStats *EventStats
		want     uint64
	}{
		{
			name: "alice 1234",
			logStats: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       1,
			},
			want: 0xbc51daf12983ca1d,
		},
		{
			name: "bob 5678",
			logStats: &EventStats{
				ServiceName: "bob",
				Fingerprint: 5678,
				Phase:       2,
			},
			want: 0x2e6e385ef8e2c36,
		},
		{
			name: "bob 567",
			logStats: &EventStats{
				ServiceName: "bob",
				Fingerprint: 567,
				Phase:       3,
			},
			want: 0xefb3faacaad3a679,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.logStats.Key()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogStats_Matches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		logStats   *EventStats
		other      stats.StatsObject
		wantResult bool
	}{
		{
			name: "matching log stats",
			logStats: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       Phase_PASSTHROUGH,
			},
			other: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       Phase_PASSTHROUGH,
			},
			wantResult: true,
		},
		{
			name: "non-matching log stats",
			logStats: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       Phase_PASSTHROUGH,
			},
			other: &EventStats{
				ServiceName: "bob",
				Fingerprint: 5678,
				Phase:       Phase_FILTERED,
			},
			wantResult: false,
		},
		{
			name: "non-log stats object",
			logStats: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       Phase_PASSTHROUGH,
			},
			other:      &stats.MockStatsObject{},
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.logStats.Matches(tt.other)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func TestLogStats_Increment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		logStats  *EventStats
		wantCount int64
	}{
		{
			name: "increment count",
			logStats: &EventStats{
				ServiceName: "alice",
				Fingerprint: 1234,
				Phase:       Phase_PASSTHROUGH,
				Count:       0,
			},
			wantCount: 0, // i = 0
		},
		{
			name: "increment count multiple times",
			logStats: &EventStats{
				ServiceName: "bob",
				Fingerprint: 5678,
				Phase:       Phase_FILTERED,
				Count:       2,
			},
			wantCount: 3, // i = 1
		},
		{
			name: "increment count again",
			logStats: &EventStats{
				ServiceName: "bob",
				Fingerprint: 5678,
				Phase:       Phase_FILTERED,
				Count:       3,
			},
			wantCount: 5, // i = 2
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.logStats.Increment("", i, int64(i))
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, tt.logStats.Count)
		})
	}
}
