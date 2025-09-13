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

func BenchmarkFingerprintingWithPools(b *testing.B) {
	fp := NewFingerprinter()
	clusterManager := NewTrieClusterManager(0.5)

	testInputs := []string{
		"INFO Received request for /api/v1/endpoint from userId=12345",
		"ERROR Failed to connect to database connection timeout after 30s",
		"DEBUG Processing user authentication for email user@example.com",
		"WARN High memory usage detected: 85% of 8GB used",
		"TRACE Executing SQL query: SELECT * FROM users WHERE active = true",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		_, _, _, _ = fp.Fingerprint(input, clusterManager)
	}
}

func BenchmarkTokenizationWithPools(b *testing.B) {
	fp := NewFingerprinter()

	testInputs := []string{
		"INFO Received request for /api/v1/endpoint from userId=12345",
		"ERROR Failed to connect to database connection timeout after 30s",
		"DEBUG Processing user authentication for email user@example.com",
		"WARN High memory usage detected: 85% of 8GB used",
		"TRACE Executing SQL query: SELECT * FROM users WHERE active = true",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testInputs[i%len(testInputs)]
		_, _, _ = fp.tokenizeString(input)
	}
}

// Benchmark object pool overhead
func BenchmarkObjectPoolOverhead(b *testing.B) {
	b.Run("TokenSeq", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ts := getTokenSeq()
			ts.add("test")
			putTokenSeq(ts)
		}
	})

	b.Run("StringSlice", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			slice := getStringSlice()
			slice = append(slice, "test")
			putStringSlice(slice)
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sb := getStringBuilder()
			sb.WriteString("test")
			_ = sb.String()
			putStringBuilder(sb)
		}
	})
}