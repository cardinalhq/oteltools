package chqpb

import (
	"testing"
	"time"

	"github.com/apache/datasketches-go/hll"
)

func TestNewMetricStatsCache(t *testing.T) {
	flushInterval := 10 * time.Minute
	cache := NewMetricStatsCache(flushInterval)

	if cache == nil {
		t.Fatalf("Expected non-nil MetricStatsCache")
	}

	if cache.flushInterval != flushInterval {
		t.Errorf("Expected flushInterval %v, got %v", flushInterval, cache.flushInterval)
	}

	if cache.hllCache == nil {
		t.Errorf("Expected non-nil hllCache map")
	}

	if cache.itemsByHour == nil {
		t.Errorf("Expected non-nil itemsByHour map")
	}
}

func testMakeUnion(t *testing.B) hll.Union {
	u, err := hll.NewUnion(12)
	if err != nil {
		t.Fatalf("Failed to create HLL union: %v", err)
	}

	if u == nil {
		t.Fatalf("Expected non-nil HLL union")
	}

	for i := int64(0); i < 1000; i++ {
		err := u.UpdateInt64(i * 4321)
		if err != nil {
			t.Fatalf("Failed to update HLL union: %v", err)
		}
	}

	return u
}

func BenchmarkGetEstimate(b *testing.B) {
	u := testMakeUnion(b)
	b.Run("GetEstimate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := u.GetEstimate()
			if err != nil {
				b.Fatalf("Failed to get estimate: %v", err)
			}
		}
	})
}

func BenchmarkToCompactSlice(b *testing.B) {
	u := testMakeUnion(b)
	b.Run("ToCompactSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := u.ToCompactSlice()
			if err != nil {
				b.Fatalf("Failed to get compact slice: %v", err)
			}
		}
	})
}

func BenchmarkUpdateString(b *testing.B) {
	u, err := hll.NewUnion(12)
	if err != nil {
		b.Fatalf("Failed to create HLL union: %v", err)
	}

	b.Run("UpdateString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := u.UpdateString("test")
			if err != nil {
				b.Fatalf("Failed to update HLL union: %v", err)
			}
		}
	})
}
