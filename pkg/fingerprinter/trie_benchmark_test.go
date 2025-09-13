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
	"testing"
)

func BenchmarkTrieClusterManager(b *testing.B) {
	fp := NewFingerprinter()
	clusterManager := NewTrieClusterManager(0.5)

	testInputs := []string{
		"INFO Processing request for user id 12345 from endpoint /api/v1/users",
		"ERROR Database connection failed timeout after 30 seconds retrying",
		"DEBUG User authentication successful for email user@example.com",
		"WARN Memory usage high 85% of 8GB available heap space used",
		"TRACE SQL query execution SELECT * FROM users WHERE active = true",
		"INFO Request completed successfully in 150ms for user 67890",
		"ERROR Failed to parse JSON payload invalid format in request body",
		"DEBUG Cache hit for key user_profile_12345 returning cached data",
		"WARN Rate limit approaching 90% of 1000 requests per minute limit",
		"TRACE HTTP response sent status 200 content-type application/json",
	}

	// Pre-tokenize inputs to isolate clustering performance
	tokenizedInputs := make([]*tokenSeq, len(testInputs))
	for i, input := range testInputs {
		ts, _, _, err := fp.testTokenizeInput(input)
		if err != nil {
			b.Fatal(err)
		}
		tokenizedInputs[i] = ts
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ts := tokenizedInputs[i%len(tokenizedInputs)]
		_ = clusterManager.cluster(ts)
	}
}

func BenchmarkJaccardSimilarity(b *testing.B) {
	set1 := map[string]struct{}{
		"info": {}, "processing": {}, "request": {}, "user": {}, "id": {},
	}
	set2 := map[string]struct{}{
		"info": {}, "request": {}, "completed": {}, "user": {}, "endpoint": {},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = jaccardSimilarity(set1, set2)
	}
}

func BenchmarkTriePoolOverhead(b *testing.B) {
	b.Run("StringSet", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m := getStringSet()
			m["test"] = struct{}{}
			putStringSet(m)
		}
	})

	b.Run("SeqNodeSlice", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := getSeqNodeSlice()
			slice = append(slice, &seqNode{})
			putSeqNodeSlice(slice)
		}
	})

	b.Run("Cluster", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c := getCluster()
			c.Fingerprint = 12345
			putCluster(c)
		}
	})
}