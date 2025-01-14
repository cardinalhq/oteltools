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

package maputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]any
		want []string
	}{
		{
			name: "flat map",
			m: map[string]any{
				"a": 1,
				"b": 2,
			},
			want: []string{"a", "b"},
		},
		{
			name: "nested map",
			m: map[string]any{
				"a": map[string]any{
					"b": 1,
					"c": 2,
				},
				"d": 3,
			},
			want: []string{"a.b", "a.c", "d"},
		},
		{
			name: "deeply nested map",
			m: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": 1,
					},
				},
				"d": 2,
			},
			want: []string{"a.b.c", "d"},
		},
		{
			name: "slice",
			m: map[string]any{
				"a": []any{1, 2},
			},
			want: []string{"a"},
		},
		{
			name: "slice of maps",
			m: map[string]any{
				"a": []any{
					map[string]any{
						"b": 1,
					},
					map[string]any{
						"c": 2,
					},
					1,
				},
			},
			want: []string{"a.0.b", "a.1.c"},
		},
		{
			name: "empty map",
			m:    map[string]any{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeepKeys(tt.m)
			assert.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkDeepKeys(b *testing.B) {
	m := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": 1,
			},
		},
		"d": 2,
		"e": []any{
			map[string]any{
				"f": 3,
			},
		},
	}

	for i := 0; i < b.N; i++ {
		DeepKeys(m)
	}
}
