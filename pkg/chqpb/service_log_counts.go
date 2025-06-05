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
	mu              sync.Mutex
	serviceName     string
	clusterName     string
	namespaceName   string
	totalCount      int64
	errorCount      int64
	exceptionCount  int64
	exceptionCounts map[int64]int64
	exceptionMap    map[int64]string
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
		entry = &logEntry{
			serviceName:     svc,
			clusterName:     clus,
			namespaceName:   ns,
			exceptionCounts: make(map[int64]int64),
			exceptionMap:    make(map[int64]string),
		}
		logMap.Store(tid, entry)
	} else {
		entry = val.(*logEntry)
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.totalCount++
	if isError(logRecord) {
		fpVal, fpFound := logRecord.Attributes().Get(translate.CardinalFieldFingerprint)
		if fpFound {
			fp := fpVal.Int()
			entry.errorCount++
			entry.exceptionCount++
			if _, exists := entry.exceptionMap[fp]; !exists {
				bytes, err := c.logToJson(logRecord, resource)
				if err == nil {
					entry.exceptionMap[fp] = string(bytes)
				}
			}
			entry.exceptionCounts[fp]++
		}
	}
}

func isError(lr plog.LogRecord) bool {
	severity := lr.SeverityNumber()
	return severity >= plog.SeverityNumberWarn
}

// flush constructs a LogSketchList from completed intervals and invokes flushFunc.
func (c *ServiceLogCountsCache) flush() {
	now := time.Now().Truncate(c.interval).Unix()
	list := &ServiceLogCountList{CustomerId: c.customerId}

	c.buckets.Range(func(intervalKey, v interface{}) bool {
		interval := intervalKey.(int64)
		if interval >= now {
			return true
		}

		logMap := v.(*sync.Map)

		logMap.Range(func(key, entryVal interface{}) bool {
			entry := entryVal.(*logEntry)

			entry.mu.Lock()
			// Build LogSketchProto for this entry
			proto := &ServiceLogCountProto{
				ServiceName:     entry.serviceName,
				NamespaceName:   entry.namespaceName,
				ClusterName:     entry.clusterName,
				Tid:             key.(string),
				Interval:        interval,
				TotalCount:      entry.totalCount,
				ErrorCount:      entry.errorCount,
				ExceptionCount:  entry.exceptionCount,
				ExceptionMap:    entry.exceptionMap,
				ExceptionCounts: entry.exceptionCounts,
			}
			list.Sketches = append(list.Sketches, proto)
			entry.mu.Unlock()

			return true
		})

		c.buckets.Delete(intervalKey)
		return true
	})

	if len(list.Sketches) > 0 {
		if err := c.flushFunc(list); err != nil {
			slog.Error("failed to flush log sketches", slog.String("customerId", c.customerId), slog.String("error", err.Error()))
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
