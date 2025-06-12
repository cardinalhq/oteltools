package chqpb

import (
	"github.com/cardinalhq/oteltools/pkg/translate"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func createTestLogRecord(ts time.Time, severity plog.SeverityNumber, body string) (pcommon.Resource, plog.LogRecord) {
	res := pcommon.NewResource()
	attrs := res.Attributes()
	attrs.PutStr(string(semconv.ServiceNameKey), "svc1")
	attrs.PutStr(string(semconv.K8SClusterNameKey), "clus1")
	attrs.PutStr(string(semconv.K8SNamespaceNameKey), "ns1")

	ld := plog.NewLogs()
	rs := ld.ResourceLogs().AppendEmpty()
	res.CopyTo(rs.Resource())

	sls := rs.ScopeLogs().AppendEmpty()
	lr := sls.LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(ts))
	lr.Attributes().PutInt(translate.CardinalFieldFingerprint, 1)

	lr.SetSeverityNumber(severity)
	lr.Body().SetStr(body)

	// Return the Resource and the LogRecord itself (the caller must copy if needed)
	return res, lr
}

func TestLogSketchCache_FlushPlainLog(t *testing.T) {
	var captured *ServiceLogCountList
	cache := NewServiceLogCountsCache(1*time.Second, "cust1", func(list *ServiceLogCountList) error {
		captured = list
		return nil
	})

	past := time.Now().Add(-2 * time.Second)
	res, lr := createTestLogRecord(past, plog.SeverityNumberInfo, "normal message")

	cache.Update(res, lr)
	cache.flush()

	if captured == nil {
		t.Fatalf("expected flush to capture a LogSketchList, got nil")
	}
	if len(captured.Sketches) != 1 {
		t.Fatalf("expected 1 LogSketchProto, got %d", len(captured.Sketches))
	}

	proto := captured.Sketches[0]
	if proto.ServiceName != "svc1" {
		t.Errorf("expected ServiceName 'svc1', got '%s'", proto.ServiceName)
	}
	if proto.NamespaceName != "ns1" {
		t.Errorf("expected NamespaceName 'ns1', got '%s'", proto.NamespaceName)
	}
	if proto.ClusterName != "clus1" {
		t.Errorf("expected ClusterName 'clus1', got '%s'", proto.ClusterName)
	}
	if proto.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", proto.TotalCount)
	}
	if proto.ErrorCount != 0 {
		t.Errorf("expected ErrorCount 0 for a non-error log, got %d", proto.ErrorCount)
	}
	if proto.ExceptionCount != 0 {
		t.Errorf("expected ExceptionCount 0 for a non-error log, got %d", proto.ExceptionCount)
	}
}

func TestLogSketchCache_FlushErrorLog(t *testing.T) {
	var captured *ServiceLogCountList
	cache := NewServiceLogCountsCache(1*time.Second, "cust1", func(list *ServiceLogCountList) error {
		captured = list
		return nil
	})

	past := time.Now().Add(-2 * time.Second)
	res, lr := createTestLogRecord(past, plog.SeverityNumberWarn, "Error occurred")

	cache.Update(res, lr)
	cache.flush()

	if captured == nil {
		t.Fatalf("expected flush to capture a LogSketchList, got nil")
	}
	if len(captured.Sketches) != 1 {
		t.Fatalf("expected 1 LogSketchProto, got %d", len(captured.Sketches))
	}

	proto := captured.Sketches[0]
	if proto.ErrorCount != 1 {
		t.Errorf("expected ErrorCount 1 for an error log, got %d", proto.ErrorCount)
	}
	if proto.ExceptionCount != 1 {
		t.Errorf("expected ExceptionCount 1 for an error log containing 'Error', got %d", proto.ExceptionCount)
	}
	if len(proto.ExceptionMap) != 1 {
		t.Errorf("expected 1 exception fingerprint, got %d", len(proto.ExceptionMap))
	}
	if len(proto.ExceptionCounts) != 1 {
		t.Errorf("expected 1 entry in ExceptionCounts, got %d", len(proto.ExceptionCounts))
	}
	for _, count := range proto.ExceptionCounts {
		if count != 1 {
			t.Errorf("expected exception count 1, got %d", count)
		}
	}
}
