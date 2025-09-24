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
	"fmt"
	"testing"
	"time"
)

var testLogSamples = []string{
	`2024-01-15 10:23:45 ERROR Failed to connect to database server at host=db.example.com port=5432`,
	`{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout after 30s","host":"api-server-1","requestId":"abc123xyz"}`,
	`INFO [2024-01-15 10:23:45] User john.doe@example.com logged in successfully from IP 192.168.1.100`,
	`WARN Processing took 1234ms which exceeds threshold of 1000ms for endpoint /api/v1/users/profile`,
	`DEBUG Query executed: SELECT * FROM users WHERE status='active' AND created_at > '2024-01-01' LIMIT 100`,
	`{"level":"info","ts":"2024-01-15T10:23:45.123Z","caller":"main.go:42","msg":"Starting application","version":"1.2.3","environment":"production"}`,
	`ERROR [RequestID: req-456-def] Payment processing failed for order #ORD-2024-001 amount=$1,234.56 currency=USD`,
	`2024-01-15 10:23:45.678 [Thread-123] WARN Memory usage at 85% (8.5GB/10GB), triggering cleanup`,
}

// Comparison benchmarks between original and optimized implementations

func BenchmarkOriginalFingerprinter(b *testing.B) {
	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testLogSamples[i%len(testLogSamples)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}
}

func BenchmarkOptimizedFingerprinter(b *testing.B) {
	fp := NewOptimizedFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testLogSamples[i%len(testLogSamples)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}
}

func BenchmarkOriginalJSON(b *testing.B) {
	jsonLog := `{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout","host":"server-1","details":{"retry":3,"timeout":30}}`

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(jsonLog, cm)
	}
}

func BenchmarkOptimizedJSON(b *testing.B) {
	jsonLog := `{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout","host":"server-1","details":{"retry":3,"timeout":30}}`

	fp := NewOptimizedFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(jsonLog, cm)
	}
}

func BenchmarkOriginalLongString(b *testing.B) {
	longLog := `ERROR [2024-01-15 10:23:45.123] [RequestID: req-abc-123-def-456] [UserID: user-789-ghi-012] Failed to process transaction for order #ORD-2024-0001234 with items: [{"sku":"PROD-A-123","qty":5,"price":99.99},{"sku":"PROD-B-456","qty":2,"price":149.99}] due to payment gateway timeout after 30000ms`

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(longLog, cm)
	}
}

func BenchmarkOptimizedLongString(b *testing.B) {
	longLog := `ERROR [2024-01-15 10:23:45.123] [RequestID: req-abc-123-def-456] [UserID: user-789-ghi-012] Failed to process transaction for order #ORD-2024-0001234 with items: [{"sku":"PROD-A-123","qty":5,"price":99.99},{"sku":"PROD-B-456","qty":2,"price":149.99}] due to payment gateway timeout after 30000ms`

	fp := NewOptimizedFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(longLog, cm)
	}
}

// Parallel benchmarks to test concurrency

func BenchmarkOriginalParallel(b *testing.B) {
	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			input := testLogSamples[i%len(testLogSamples)]
			_, _, _, _ = fp.Fingerprint(input, cm)
			i++
		}
	})
}

func BenchmarkOptimizedParallel(b *testing.B) {
	fp := NewOptimizedFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			input := testLogSamples[i%len(testLogSamples)]
			_, _, _, _ = fp.Fingerprint(input, cm)
			i++
		}
	})
}

// Test memory allocation patterns over time
func TestMemoryPattern(t *testing.T) {
	iterations := 100000
	samples := []string{
		`ERROR Database connection failed`,
		`{"level":"error","message":"Timeout occurred","duration":1234}`,
		`INFO Successfully processed 1000 records in 45.3 seconds`,
	}

	// Test original implementation
	t.Run("Original", func(t *testing.T) {
		fp := NewFingerprinter()
		cm := NewTrieClusterManager(0.8)

		start := time.Now()
		for i := 0; i < iterations; i++ {
			input := samples[i%len(samples)]
			_, _, _, _ = fp.Fingerprint(input, cm)
		}
		duration := time.Since(start)

		t.Logf("Original: %d iterations in %v (%v per op)",
			iterations, duration, duration/time.Duration(iterations))
	})

	// Test optimized implementation
	t.Run("Optimized", func(t *testing.T) {
		fp := NewOptimizedFingerprinter()
		cm := NewTrieClusterManager(0.8)

		start := time.Now()
		for i := 0; i < iterations; i++ {
			input := samples[i%len(samples)]
			_, _, _, _ = fp.Fingerprint(input, cm)
		}
		duration := time.Since(start)

		t.Logf("Optimized: %d iterations in %v (%v per op)",
			iterations, duration, duration/time.Duration(iterations))
	})
}

// Benchmark suite runner
func TestBenchmarkSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark suite in short mode")
	}

	tests := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"Original", BenchmarkOriginalFingerprinter},
		{"Optimized", BenchmarkOptimizedFingerprinter},
		{"OriginalJSON", BenchmarkOriginalJSON},
		{"OptimizedJSON", BenchmarkOptimizedJSON},
		{"OriginalLong", BenchmarkOriginalLongString},
		{"OptimizedLong", BenchmarkOptimizedLongString},
	}

	for _, test := range tests {
		result := testing.Benchmark(test.fn)
		fmt.Printf("%-15s: %10d ns/op %8d B/op %5d allocs/op\n",
			test.name,
			result.NsPerOp(),
			result.AllocedBytesPerOp(),
			result.AllocsPerOp())
	}
}