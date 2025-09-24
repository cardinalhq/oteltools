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
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
)

func TestOptimizedMemoryProfile(t *testing.T) {
	// Create output directory for profiles
	os.MkdirAll("profiles", 0755)

	testLogs := []string{
		`2024-01-15 10:23:45 ERROR Failed to connect to database server`,
		`{"level":"error","message":"Connection timeout","host":"server-1"}`,
		`INFO User logged in successfully`,
		`WARN Processing took 1234ms`,
		`DEBUG Query executed successfully`,
	}

	fp := NewOptimizedFingerprinter()
	cm := NewTrieClusterManager(0.8)

	// Start profiling
	heapFile, err := os.Create("profiles/optimized_heap.prof")
	if err != nil {
		t.Fatalf("Could not create heap profile: %v", err)
	}
	defer heapFile.Close()

	// Warm up
	for i := 0; i < 100; i++ {
		input := testLogs[i%len(testLogs)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}

	// Force GC and reset stats
	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)

	// Run test
	iterations := 10000
	for i := 0; i < iterations; i++ {
		input := testLogs[i%len(testLogs)]
		_, _, _, _ = fp.Fingerprint(input, cm)
	}

	// Get final stats
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Write heap profile
	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		t.Fatalf("Could not write heap profile: %v", err)
	}

	// Report
	t.Logf("Optimized Memory Profile:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Allocations per op: %d", (m1.Mallocs-m0.Mallocs)/uint64(iterations))
	t.Logf("  Bytes per op: %d", (m1.TotalAlloc-m0.TotalAlloc)/uint64(iterations))
	t.Logf("  Total allocations: %d", m1.Mallocs-m0.Mallocs)
	t.Logf("  Total bytes: %d", m1.TotalAlloc-m0.TotalAlloc)
	t.Logf("\nProfile saved to profiles/optimized_heap.prof")
	t.Logf("Check with: go tool pprof -top profiles/optimized_heap.prof")
}