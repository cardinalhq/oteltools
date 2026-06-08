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

package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otellogapi "go.opentelemetry.io/otel/log"
)

// otlpRecorder is an httptest server that counts inbound OTLP/HTTP POSTs and
// returns the empty success body the exporters expect.
func otlpRecorder(t *testing.T) (*httptest.Server, *int64) {
	t.Helper()
	var count int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt64(&count, 1)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{})
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

func TestWithSecondaryExporter_SetsConfig(t *testing.T) {
	cfg := &setupConfig{}
	WithSecondaryExporter("https://example.com:4318", map[string]string{"x": "y"})(cfg)
	require.NotNil(t, cfg.secondary)
	assert.Equal(t, "https://example.com:4318", cfg.secondary.endpoint)
	assert.Equal(t, map[string]string{"x": "y"}, cfg.secondary.headers)
}

func TestWithSecondaryExporter_EmptyEndpointIsNoop(t *testing.T) {
	cfg := &setupConfig{}
	WithSecondaryExporter("", map[string]string{"x": "y"})(cfg)
	assert.Nil(t, cfg.secondary)
}

func TestTracerProvider_FansOutToSecondary(t *testing.T) {
	primary, primaryCount := otlpRecorder(t)
	secondary, secondaryCount := otlpRecorder(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", primary.URL)

	ctx := context.Background()
	tp, err := newTracerProvider(ctx, true, nil, nil, &secondaryExporter{endpoint: secondary.URL})
	require.NoError(t, err)

	_, span := tp.Tracer("test").Start(ctx, "span")
	span.End()
	require.NoError(t, tp.ForceFlush(ctx))
	require.NoError(t, tp.Shutdown(ctx))

	assert.Positive(t, atomic.LoadInt64(primaryCount), "primary must receive traces")
	assert.Positive(t, atomic.LoadInt64(secondaryCount), "secondary must receive traces")
}

func TestTracerProvider_NoSecondaryOnlyPrimary(t *testing.T) {
	primary, primaryCount := otlpRecorder(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", primary.URL)

	ctx := context.Background()
	tp, err := newTracerProvider(ctx, true, nil, nil, nil)
	require.NoError(t, err)

	_, span := tp.Tracer("test").Start(ctx, "span")
	span.End()
	require.NoError(t, tp.ForceFlush(ctx))
	require.NoError(t, tp.Shutdown(ctx))

	assert.Positive(t, atomic.LoadInt64(primaryCount), "primary must receive traces")
}

func TestMeterProvider_FansOutToSecondary(t *testing.T) {
	primary, primaryCount := otlpRecorder(t)
	secondary, secondaryCount := otlpRecorder(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", primary.URL)

	ctx := context.Background()
	mp, err := newMeterProvider(ctx, true, nil, nil, &secondaryExporter{endpoint: secondary.URL})
	require.NoError(t, err)

	counter, err := mp.Meter("test").Int64Counter("c")
	require.NoError(t, err)
	counter.Add(ctx, 1)
	require.NoError(t, mp.ForceFlush(ctx))
	require.NoError(t, mp.Shutdown(ctx))

	assert.Positive(t, atomic.LoadInt64(primaryCount), "primary must receive metrics")
	assert.Positive(t, atomic.LoadInt64(secondaryCount), "secondary must receive metrics")
}

func TestLoggerProvider_FansOutToSecondary(t *testing.T) {
	primary, primaryCount := otlpRecorder(t)
	secondary, secondaryCount := otlpRecorder(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", primary.URL)

	ctx := context.Background()
	lp, err := newLoggerProvider(ctx, true, nil, nil, &secondaryExporter{endpoint: secondary.URL})
	require.NoError(t, err)

	var rec otellogapi.Record
	rec.SetBody(otellogapi.StringValue("hello"))
	lp.Logger("test").Emit(ctx, rec)
	require.NoError(t, lp.ForceFlush(ctx))
	require.NoError(t, lp.Shutdown(ctx))

	assert.Positive(t, atomic.LoadInt64(primaryCount), "primary must receive logs")
	assert.Positive(t, atomic.LoadInt64(secondaryCount), "secondary must receive logs")
}
