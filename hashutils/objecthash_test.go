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

package hashutils

import (
	"hash/fnv"
	"testing"

	"github.com/cespare/xxhash"
)

func TestHashAny(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  uint64
	}{
		{"nil", nil, 3078252789000966458},
		{"bool true", true, 12201293588068846061},
		{"bool false", false, 14672055408725111526},
		{"float64", 3.14, 7703445615302545898},
		{"string", "hello", 8843100075938556332},
		{"empty array", []any{}, 5746651081700789484},
		{"array", []any{1, "two", 3.0}, 14901291645572418406},
		{"empty map", map[string]any{}, 13407604179352847978},
		{"map", map[string]any{"one": 1, "two": "two"}, 8190209883385698026},
		{"map with swapped order", map[string]any{"two": "two", "one": 1}, 8190209883385698026},
		{"nested map", map[string]any{"one": 1, "two": map[string]any{"three": 3}}, 15507829625135637265},
		{"struct", struct{ A int }{1}, 14553743525572530200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := fnv.New64a()
			got := HashAny(tt.value, hasher)
			if got != tt.want {
				t.Errorf("HashAny() = %v, want %v", got, tt.want)
			}
		})
	}
}
func BenchmarkHashAnyFNV(b *testing.B) {
	value := map[string]any{"one": 1, "two": "two"}
	for i := 0; i < b.N; i++ {
		hasher := fnv.New64a()
		HashAny(value, hasher)
	}
}

func BenchmarkHashAnyXXHash(b *testing.B) {
	value := map[string]any{"one": 1, "two": "two"}
	for i := 0; i < b.N; i++ {
		hasher := xxhash.New()
		HashAny(value, hasher)
	}
}
