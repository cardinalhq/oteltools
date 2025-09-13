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
	"time"

	"github.com/cardinalhq/oteltools/pkg/fingerprinter/tokenizer"
)

// Object pools to reduce GC pressure from frequent allocations

var (
	// Pool for tokenSeq objects
	tokenSeqPool = sync.Pool{
		New: func() interface{} {
			return &tokenSeq{
				items:    make([]string, 0, 16), // Pre-allocate some capacity
				jsonKeys: make([]string, 0, 8),
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

	// Pool for map[string]struct{} used in token sets
	stringSetPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]struct{}, 16)
		},
	}

	// Pool for []*seqNode slices used in collectLeafers
	seqNodeSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]*seqNode, 0, 8)
		},
	}

	// Pool for cluster structs
	clusterPool = sync.Pool{
		New: func() interface{} {
			return &cluster{}
		},
	}
)

// getTokenSeq gets a tokenSeq from the pool and resets it
func getTokenSeq() *tokenSeq {
	ts := tokenSeqPool.Get().(*tokenSeq)
	ts.index = 0
	ts.items = ts.items[:0]       // Reset slice length but keep capacity
	ts.jsonKeys = ts.jsonKeys[:0] // Reset slice length but keep capacity
	return ts
}

// putTokenSeq returns a tokenSeq to the pool
func putTokenSeq(ts *tokenSeq) {
	// Don't pool extremely large slices to avoid memory bloat
	if cap(ts.items) > 256 || cap(ts.jsonKeys) > 64 {
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
	//nolint:staticcheck // SA6002: slice allocation is acceptable for pooling
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

// getStringSet gets a string set from the pool and clears it
func getStringSet() map[string]struct{} {
	m := stringSetPool.Get().(map[string]struct{})
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// putStringSet returns a string set to the pool
func putStringSet(m map[string]struct{}) {
	// Don't pool extremely large maps
	if len(m) > 128 {
		return
	}
	stringSetPool.Put(m)
}

// getSeqNodeSlice gets a seqNode slice from the pool and resets it
func getSeqNodeSlice() []*seqNode {
	slice := seqNodeSlicePool.Get().([]*seqNode)
	return slice[:0] // Reset length but keep capacity
}

// putSeqNodeSlice returns a seqNode slice to the pool
func putSeqNodeSlice(slice []*seqNode) {
	// Don't pool extremely large slices
	if cap(slice) > 64 {
		return
	}
	//nolint:staticcheck // SA6002: slice allocation is acceptable for pooling
	seqNodeSlicePool.Put(slice)
}

// getCluster gets a cluster from the pool and resets it
func getCluster() *cluster {
	c := clusterPool.Get().(*cluster)
	// Reset the cluster
	c.Fingerprint = 0
	c.TokenSet = nil
	c.MatchCount = 0
	c.Total = 0
	c.LastUpdated = time.Time{}
	return c
}

// putCluster returns a cluster to the pool
func putCluster(c *cluster) {
	clusterPool.Put(c)
}