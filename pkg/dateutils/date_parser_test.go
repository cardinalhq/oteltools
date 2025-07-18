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

package dateutils

import (
	"testing"
	"time"
)

func TestToStartEnd(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name     string
		startStr string
		endStr   string
		wantErr  bool
		validate func(start, end int64) bool
	}{
		{
			name:     "default empty input uses e-1h to now",
			startStr: "",
			endStr:   "",
			wantErr:  false,
			validate: func(start, end int64) bool {
				return end > start && end-start >= int64(time.Hour/time.Millisecond)
			},
		},
		{
			name:     "fixed ISO8601 start and now end",
			startStr: "2024-01-01T00:00:00Z",
			endStr:   "now",
			wantErr:  false,
			validate: func(start, end int64) bool {
				expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
				return start == expectedStart && end > expectedStart
			},
		},
		{
			name:     "relative start, fixed end",
			startStr: "e-2h",
			endStr:   now.Format(time.RFC3339),
			wantErr:  false,
			validate: func(start, end int64) bool {
				return end-start == int64(2*time.Hour/time.Millisecond)
			},
		},
		{
			name:     "fixed start, relative end",
			startStr: now.Add(-3 * time.Hour).Format(time.RFC3339),
			endStr:   "e+3h",
			wantErr:  false,
			validate: func(start, end int64) bool {
				return end-start == int64(3*time.Hour/time.Millisecond)
			},
		},
		{
			name:     "relative start and end with proper delta",
			startStr: "e-1h",
			endStr:   "now-30m",
			wantErr:  false,
			validate: func(start, end int64) bool {
				expectedDelta := int64(time.Hour / time.Millisecond)
				actualDelta := end - start
				const leeway = int64(2000) // 2s wiggle room for clock skew
				return actualDelta-expectedDelta <= leeway
			},
		},
		{
			name:     "invalid date format",
			startStr: "banana",
			endStr:   "now",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ToStartEnd(tt.startStr, tt.endStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStartEnd() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.validate != nil {
				if !tt.validate(start, end) {
					t.Errorf("ToStartEnd() start/end validation failed: got start=%d, end=%d", start, end)
				}
			}
		})
	}
}
