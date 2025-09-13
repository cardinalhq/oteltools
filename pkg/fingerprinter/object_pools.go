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
	"strings"
	"sync"

	"github.com/cardinalhq/oteltools/pkg/fingerprinter/tokenizer"
)

// Object pools to reduce GC pressure from frequent allocations

var (
	// Pool for TokenSeq objects
	tokenSeqPool = sync.Pool{
		New: func() interface{} {
			return &TokenSeq{
				Items:    make([]string, 0, 16), // Pre-allocate some capacity
				JSONKeys: make([]string, 0, 8),
			}
		},
	}

	// Pool for string slices used in quoted strings processing
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 4)
		},
	}

	// Pool for string builders
	stringBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}

	// Pool for tokenizers (though they're cheap to create)
	tokenizerPool = sync.Pool{
		New: func() interface{} {
			return tokenizer.NewFingerprintTokenizer()
		},
	}
)

// getTokenSeq gets a TokenSeq from the pool and resets it
func getTokenSeq() *TokenSeq {
	ts := tokenSeqPool.Get().(*TokenSeq)
	ts.index = 0
	ts.Items = ts.Items[:0]     // Reset slice length but keep capacity
	ts.JSONKeys = ts.JSONKeys[:0] // Reset slice length but keep capacity
	return ts
}

// putTokenSeq returns a TokenSeq to the pool
func putTokenSeq(ts *TokenSeq) {
	// Don't pool extremely large slices to avoid memory bloat
	if cap(ts.Items) > 256 || cap(ts.JSONKeys) > 64 {
		return
	}
	tokenSeqPool.Put(ts)
}

// getStringSlice gets a string slice from the pool and resets it
func getStringSlice() []string {
	slice := stringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// putStringSlice returns a string slice to the pool
func putStringSlice(slice []string) {
	// Don't pool extremely large slices
	if cap(slice) > 64 {
		return
	}
	stringSlicePool.Put(slice)
}

// getStringBuilder gets a string builder from the pool and resets it
func getStringBuilder() *strings.Builder {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// putStringBuilder returns a string builder to the pool
func putStringBuilder(sb *strings.Builder) {
	// Don't pool extremely large builders to avoid memory bloat
	if sb.Cap() > 1024 {
		return
	}
	stringBuilderPool.Put(sb)
}

// getTokenizer gets a tokenizer from the pool
func getTokenizer() *tokenizer.FingerprintTokenizer {
	return tokenizerPool.Get().(*tokenizer.FingerprintTokenizer)
}

// putTokenizer returns a tokenizer to the pool
func putTokenizer(tk *tokenizer.FingerprintTokenizer) {
	tokenizerPool.Put(tk)
}