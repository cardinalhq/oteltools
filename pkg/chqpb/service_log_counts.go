// Copyright 2024-2025 CardinalHQ, Inc.
//
// CardinalHQ, Inc. proprietary and confidential.
// Unauthorized copying, distribution, or modification of this file,
// via any medium, is strictly prohibited without prior written consent.
//
// All rights reserved.

package chqpb

import (
	"fmt"
	"github.com/cardinalhq/oteltools/pkg/translate"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"hash/fnv"
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"golang.org/x/exp/slog"
)

type logEntry struct {
	mu            sync.Mutex
	serviceName   string
	clusterName   string
	namespaceName string
	proto         *ServiceLogCountProto
}

// ServiceLogCountsCache aggregates log metrics into time buckets, then flushes as LogSketchList.
type ServiceLogCountsCache struct {
	buckets    sync.Map // map[int64]*sync.Map where inner map: map[string]*logEntry
	customerId string
	interval   time.Duration
	flushFunc  func(list *ServiceLogCountList) error
	marshaller plog.JSONMarshaler
}

// NewServiceLogCountsCache creates a cache that flushes every interval, invoking flushFunc with LogSketchList.
func NewServiceLogCountsCache(interval time.Duration, cid string, flushFunc func(list *ServiceLogCountList) error) *ServiceLogCountsCache {
	c := &ServiceLogCountsCache{
		interval:   interval,
		customerId: cid,
		flushFunc:  flushFunc,
		marshaller: plog.JSONMarshaler{},
	}
	go c.loop()
	return c
}

func (c *ServiceLogCountsCache) loop() {
	t := time.NewTicker(c.interval)
	for range t.C {
		c.flush()
	}
}

// Update ingests a single logRecord with resource attributes.
func (c *ServiceLogCountsCache) Update(resource pcommon.Resource, logRecord plog.LogRecord) {
	// Compute interval bucket
	interval := logRecord.Timestamp().AsTime().Truncate(c.interval).Unix()

	// Extract service, cluster, namespace
	svcAttr, svcFound := resource.Attributes().Get(string(semconv.ServiceNameKey))
	clAttr, clFound := resource.Attributes().Get(string(semconv.K8SClusterNameKey))
	nsAttr, nsFound := resource.Attributes().Get(string(semconv.K8SNamespaceNameKey))

	svc := ""
	if svcFound {
		svc = svcAttr.AsString()
	}
	clus := ""
	if clFound {
		clus = clAttr.AsString()
	}
	ns := ""
	if nsFound {
		ns = nsAttr.AsString()
	}

	if svc == "" || clus == "" || ns == "" {
		return
	}

	// Build TID as "service|cluster|namespace"
	h := fnv.New64a()
	tidStr := fmt.Sprintf("%s|%s|%s", svc, clus, ns)
	_, _ = h.Write([]byte(tidStr))
	tid := fmt.Sprintf("%x", h.Sum64())

	// Lookup or create entry in this interval bucket
	bucketIface, _ := c.buckets.LoadOrStore(interval, &sync.Map{})
	logMap := bucketIface.(*sync.Map)

	val, ok := logMap.Load(tid)
	var entry *logEntry
	if !ok {
		proto := &ServiceLogCountProto{
			ServiceName:     svc,
			NamespaceName:   ns,
			ClusterName:     clus,
			Tid:             tid,
			Interval:        interval,
			TotalCount:      0,
			ErrorCount:      0,
			ExceptionCount:  0,
			ExceptionMap:    make(map[int64]string),
			ExceptionCounts: make(map[int64]int64),
		}
		entry = &logEntry{
			serviceName:   svc,
			clusterName:   clus,
			namespaceName: ns,
			proto:         proto,
		}
		logMap.Store(tid, entry)
	} else {
		entry = val.(*logEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.proto.TotalCount++
	if isError(logRecord) {
		fpVal, fpFound := logRecord.Attributes().Get(translate.CardinalFieldFingerprint)
		if fpFound {
			fp := fpVal.Int()
			entry.proto.ErrorCount++
			entry.proto.ExceptionCount++ // only bump for first‐time fp
			if _, exists := entry.proto.ExceptionMap[fp]; !exists {
				bytes, err := c.logToJson(logRecord, resource)
				if err == nil {
					entry.proto.ExceptionMap[fp] = string(bytes)
				}
			}
			entry.proto.ExceptionCounts[fp]++
		}
	}
}

func isError(lr plog.LogRecord) bool {
	severity := lr.SeverityNumber()
	return severity >= plog.SeverityNumberWarn
}

func (c *ServiceLogCountsCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	list := &ServiceLogCountList{CustomerId: c.customerId}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}

		skMap := v.(*sync.Map)

		skMap.Range(func(tid, entryVal interface{}) bool {
			entry := entryVal.(*logEntry)

			entry.mu.Lock()
			list.Sketches = append(list.Sketches, entry.proto)
			entry.mu.Unlock()

			return true
		})
		c.buckets.Delete(intervalKey)
		return true
	})

	if len(list.Sketches) > 0 {
		if err := c.flushFunc(list); err != nil {
			slog.Error("failed to flush service log counts", slog.String("customerId", c.customerId), slog.String("error", err.Error()))
		}
	}
}

func (c *ServiceLogCountsCache) logToJson(logRecord plog.LogRecord, resource pcommon.Resource) ([]byte, error) {
	ld := plog.NewLogs()
	rs := ld.ResourceLogs().AppendEmpty()
	resource.CopyTo(rs.Resource())

	sl := rs.ScopeLogs().AppendEmpty()
	logRecord.CopyTo(sl.LogRecords().AppendEmpty())

	return c.marshaller.MarshalLogs(ld)
}
