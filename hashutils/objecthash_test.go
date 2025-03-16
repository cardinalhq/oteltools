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
		{"nil", nil, 2397808468787316396},
		{"map", map[string]any{"one": 1, "two": "two"}, 9120657192212919766},
		{"map with swapped order", map[string]any{"two": "two", "one": 1}, 9120657192212919766},
		{"nested map", map[string]any{"one": 1, "two": map[string]any{"three": 3}}, 15256681109419306243},
		{"struct1", struct{ A int }{1}, 15276513887743359300},
		{"struct0", struct{ A int }{0}, 17508829294710948709},
		{"struct with pointers to other objects", struct{ A *int }{func() *int { i := 1; return &i }()}, 6485897241875398280},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := fnv.New64a()
			got := HashAny(hasher, tt.value)
			if got != tt.want {
				t.Errorf("HashAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultHasher(t *testing.T) {
	got := HashAny(nil, []byte("hello"))
	if got == 0 {
		t.Errorf("HashAny() = 0, want non-zero")
	}
}

func TestHashStrings(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   uint64
	}{
		{"empty", []string{}, 14695981039346656037},
		{"single string", []string{"hello"}, 12230803299529341361},
		{"multiple strings", []string{"hello", "world"}, 18168791734189485541},
		{"strings with special characters", []string{"hello", "world", "!@#$%^&*()"}, 16906879461096828933},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := fnv.New64a()
			got := HashStrings(hasher, tt.values...)
			if got != tt.want {
				t.Errorf("HashStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkHashAnyFNV(b *testing.B) {
	value := map[string]any{"one": 1, "two": "two"}
	for b.Loop() {
		hasher := fnv.New64a()
		HashAny(hasher, value)
	}
}

func BenchmarkHashAnyXXHash(b *testing.B) {
	value := map[string]any{"one": 1, "two": "two"}
	for b.Loop() {
		hasher := xxhash.New()
		HashAny(hasher, value)
	}
}
