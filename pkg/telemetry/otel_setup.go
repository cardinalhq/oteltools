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
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, opts ...Option) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	insecure := os.Getenv("OTEL_INSECURE") == "true"

	cfg := &setupConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	httpClient := cfg.resolveHTTPClient()

	// Build the merged resource: start with SDK defaults (telemetry.sdk.*,
	// host.*, process.runtime.*, OTEL_RESOURCE_ATTRIBUTES), then merge
	// caller-provided attrs on top so they take precedence on conflict.
	res := cfg.resource
	if res != nil {
		defaultRes, resErr := resource.New(ctx,
			resource.WithFromEnv(),
			resource.WithTelemetrySDK(),
			resource.WithHost(),
			resource.WithProcessRuntimeDescription(),
		)
		if resErr != nil {
			slog.Warn("failed to build default OTel resource, using caller resource only", slog.Any("error", resErr))
		} else {
			merged, mergeErr := resource.Merge(defaultRes, res)
			if mergeErr != nil {
				slog.Warn("failed to merge OTel resources, using caller resource only", slog.Any("error", mergeErr))
			} else {
				res = merged
			}
		}
	}

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		slog.Info("Shutting down OpenTelemetry SDK")
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		slog.Info("OpenTelemetry SDK shutdown complete")
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx, insecure, httpClient, res, cfg.secondary)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, insecure, httpClient, res, cfg.secondary)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, insecure, httpClient, res, cfg.secondary)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, insecure bool, httpClient *http.Client, res *resource.Resource, secondary *secondaryExporter) (*trace.TracerProvider, error) {
	opts := []otlptracehttp.Option{}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if httpClient != nil {
		opts = append(opts, otlptracehttp.WithHTTPClient(httpClient))
	}
	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	tpOpts := []trace.TracerProviderOption{trace.WithBatcher(traceExporter)}
	if secondary != nil {
		secOpts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(secondary.endpoint)}
		if len(secondary.headers) > 0 {
			secOpts = append(secOpts, otlptracehttp.WithHeaders(secondary.headers))
		}
		secExporter, err := otlptracehttp.New(ctx, secOpts...)
		if err != nil {
			return nil, err
		}
		tpOpts = append(tpOpts, trace.WithBatcher(secExporter))
	}
	if res != nil {
		tpOpts = append(tpOpts, trace.WithResource(res))
	}
	return trace.NewTracerProvider(tpOpts...), nil
}

func newMeterProvider(ctx context.Context, insecure bool, httpClient *http.Client, res *resource.Resource, secondary *secondaryExporter) (*metric.MeterProvider, error) {
	opts := []otlpmetrichttp.Option{}
	if insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if httpClient != nil {
		opts = append(opts, otlpmetrichttp.WithHTTPClient(httpClient))
	}
	metricExporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	mpOpts := []metric.Option{metric.WithReader(metric.NewPeriodicReader(metricExporter))}
	if secondary != nil {
		secOpts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpointURL(secondary.endpoint)}
		if len(secondary.headers) > 0 {
			secOpts = append(secOpts, otlpmetrichttp.WithHeaders(secondary.headers))
		}
		secExporter, err := otlpmetrichttp.New(ctx, secOpts...)
		if err != nil {
			return nil, err
		}
		mpOpts = append(mpOpts, metric.WithReader(metric.NewPeriodicReader(secExporter)))
	}
	if res != nil {
		mpOpts = append(mpOpts, metric.WithResource(res))
	}
	return metric.NewMeterProvider(mpOpts...), nil
}

func newLoggerProvider(ctx context.Context, insecure bool, httpClient *http.Client, res *resource.Resource, secondary *secondaryExporter) (*otellog.LoggerProvider, error) {
	opts := []otlploghttp.Option{}
	if insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if httpClient != nil {
		opts = append(opts, otlploghttp.WithHTTPClient(httpClient))
	}
	logExporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	lpOpts := []otellog.LoggerProviderOption{otellog.WithProcessor(otellog.NewBatchProcessor(logExporter))}
	if secondary != nil {
		secOpts := []otlploghttp.Option{otlploghttp.WithEndpointURL(secondary.endpoint)}
		// otlploghttp's WithEndpointURL uses the URL path verbatim; a path-less
		// endpoint resolves to "/" (404 at an OTLP collector) rather than the
		// standard "/v1/logs". The otlptracehttp/otlpmetrichttp exporters default
		// an empty path to /v1/{signal}, but the log exporter does not — so the
		// secondary log sink silently 404s while metrics/traces flow. Force the
		// standard path when the operator-supplied endpoint carries none.
		if u, perr := url.Parse(secondary.endpoint); perr == nil && (u.Path == "" || u.Path == "/") {
			secOpts = append(secOpts, otlploghttp.WithURLPath("/v1/logs"))
		}
		if len(secondary.headers) > 0 {
			secOpts = append(secOpts, otlploghttp.WithHeaders(secondary.headers))
		}
		secExporter, err := otlploghttp.New(ctx, secOpts...)
		if err != nil {
			return nil, err
		}
		lpOpts = append(lpOpts, otellog.WithProcessor(otellog.NewBatchProcessor(secExporter)))
	}
	if res != nil {
		lpOpts = append(lpOpts, otellog.WithResource(res))
	}
	return otellog.NewLoggerProvider(lpOpts...), nil
}
