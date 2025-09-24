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
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

func TestMemoryProfile(t *testing.T) {
	// Create output directory for profiles
	os.MkdirAll("profiles", 0755)

	// Sample log messages with various patterns
	testLogs := []string{
		`2024-01-15 10:23:45 ERROR Failed to connect to database server at host=db.example.com port=5432`,
		`{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout after 30s","host":"api-server-1","requestId":"abc123xyz"}`,
		`INFO [2024-01-15 10:23:45] User john.doe@example.com logged in successfully from IP 192.168.1.100`,
		`WARN Processing took 1234ms which exceeds threshold of 1000ms for endpoint /api/v1/users/profile`,
		`DEBUG Query executed: SELECT * FROM users WHERE status='active' AND created_at > '2024-01-01' LIMIT 100`,
		`{"level":"info","ts":"2024-01-15T10:23:45.123Z","caller":"main.go:42","msg":"Starting application","version":"1.2.3","environment":"production"}`,
		`ERROR [RequestID: req-456-def] Payment processing failed for order #ORD-2024-001 amount=$1,234.56 currency=USD`,
		`2024-01-15 10:23:45.678 [Thread-123] WARN Memory usage at 85% (8.5GB/10GB), triggering cleanup`,
		`{"timestamp":"2024-01-15T10:23:45Z","severity":"ERROR","message":"Failed to parse JSON","error":"unexpected token at position 42","input":"{malformed json}"}`,
		`INFO Successfully processed batch job: imported 10000 records, updated 5000 records, deleted 100 records in 45.3 seconds`,
	}

	// More diverse test data
	for i := 0; i < 100; i++ {
		testLogs = append(testLogs, fmt.Sprintf(`ERROR Database query failed: timeout after %dms for query ID %d`, i*100, i))
		testLogs = append(testLogs, fmt.Sprintf(`{"level":"info","user_id":%d,"action":"clicked_button","button_id":"btn_%d","timestamp":"%s"}`, i, i, time.Now().Format(time.RFC3339)))
	}

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	// Start CPU profiling
	cpuFile, err := os.Create("profiles/fingerprinter_cpu.prof")
	if err != nil {
		t.Fatalf("Could not create CPU profile: %v", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		t.Fatalf("Could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Initial memory stats
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)
	runtime.GC()

	// Run fingerprinting in a loop to generate allocations
	iterations := 10000
	for i := 0; i < iterations; i++ {
		input := testLogs[i%len(testLogs)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}

	// Force GC and get memory stats
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Write heap profile
	heapFile, err := os.Create("profiles/fingerprinter_heap.prof")
	if err != nil {
		t.Fatalf("Could not create heap profile: %v", err)
	}
	defer heapFile.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		t.Fatalf("Could not write heap profile: %v", err)
	}

	// Write alloc profile
	allocFile, err := os.Create("profiles/fingerprinter_alloc.prof")
	if err != nil {
		t.Fatalf("Could not create alloc profile: %v", err)
	}
	defer allocFile.Close()

	if err := pprof.Lookup("allocs").WriteTo(allocFile, 0); err != nil {
		t.Fatalf("Could not write alloc profile: %v", err)
	}

	// Analyze and report memory usage
	report := map[string]interface{}{
		"iterations":        iterations,
		"initial_heap":      m0.HeapAlloc,
		"final_heap":        m1.HeapAlloc,
		"heap_increase":     m1.HeapAlloc - m0.HeapAlloc,
		"total_alloc":       m1.TotalAlloc - m0.TotalAlloc,
		"mallocs":           m1.Mallocs - m0.Mallocs,
		"frees":             m1.Frees - m0.Frees,
		"heap_objects":      m1.HeapObjects - m0.HeapObjects,
		"gc_cycles":         m1.NumGC - m0.NumGC,
		"allocs_per_op":     (m1.Mallocs - m0.Mallocs) / uint64(iterations),
		"bytes_per_op":      (m1.TotalAlloc - m0.TotalAlloc) / uint64(iterations),
		"pause_total_ns":    m1.PauseTotalNs - m0.PauseTotalNs,
		"heap_sys":          m1.HeapSys,
		"heap_idle":         m1.HeapIdle,
		"heap_in_use":       m1.HeapInuse,
		"heap_released":     m1.HeapReleased,
		"stack_in_use":      m1.StackInuse,
		"stack_sys":         m1.StackSys,
	}

	// Write report to JSON
	reportFile, err := os.Create("profiles/memory_report.json")
	if err != nil {
		t.Fatalf("Could not create report file: %v", err)
	}
	defer reportFile.Close()

	encoder := json.NewEncoder(reportFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		t.Fatalf("Could not write report: %v", err)
	}

	// Print summary
	t.Logf("Memory Profile Summary:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Allocations per operation: %d", (m1.Mallocs-m0.Mallocs)/uint64(iterations))
	t.Logf("  Bytes per operation: %d", (m1.TotalAlloc-m0.TotalAlloc)/uint64(iterations))
	t.Logf("  Total allocations: %d", m1.Mallocs-m0.Mallocs)
	t.Logf("  Total bytes allocated: %d", m1.TotalAlloc-m0.TotalAlloc)
	t.Logf("  Heap increase: %d bytes", m1.HeapAlloc-m0.HeapAlloc)
	t.Logf("  GC cycles: %d", m1.NumGC-m0.NumGC)
	t.Logf("\nProfiles written to profiles/ directory")
	t.Logf("  View with: go tool pprof profiles/fingerprinter_heap.prof")
	t.Logf("  For allocations: go tool pprof -alloc_space profiles/fingerprinter_heap.prof")
	t.Logf("  For alloc objects: go tool pprof -alloc_objects profiles/fingerprinter_alloc.prof")
}

// BenchmarkFingerprintMemory benchmarks memory allocations
func BenchmarkFingerprintMemory(b *testing.B) {
	testLogs := []string{
		`2024-01-15 10:23:45 ERROR Failed to connect to database server at host=db.example.com port=5432`,
		`{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout after 30s","host":"api-server-1"}`,
		`INFO User logged in successfully from IP 192.168.1.100`,
		`WARN Processing took 1234ms which exceeds threshold`,
		`DEBUG Query executed: SELECT * FROM users WHERE status='active'`,
	}

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		input := testLogs[i%len(testLogs)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}
}

// BenchmarkFingerprintWithJSON benchmarks JSON parsing scenarios
func BenchmarkFingerprintWithJSON(b *testing.B) {
	jsonLog := `{"timestamp":"2024-01-15T10:23:45Z","level":"error","message":"Connection timeout","host":"server-1","details":{"retry":3,"timeout":30}}`

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(jsonLog, cm)
	}
}

// BenchmarkFingerprintLongStrings benchmarks with longer input strings
func BenchmarkFingerprintLongStrings(b *testing.B) {
	longLog := `ERROR [2024-01-15 10:23:45.123] [RequestID: req-abc-123-def-456] [UserID: user-789-ghi-012] Failed to process transaction for order #ORD-2024-0001234 with items: [{"sku":"PROD-A-123","qty":5,"price":99.99},{"sku":"PROD-B-456","qty":2,"price":149.99},{"sku":"PROD-C-789","qty":1,"price":299.99}] due to payment gateway timeout after 30000ms. Retrying with exponential backoff. Current retry attempt: 3/5. Next retry in: 8000ms`

	fp := NewFingerprinter()
	cm := NewTrieClusterManager(0.8)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = fp.Fingerprint(longLog, cm)
	}
}