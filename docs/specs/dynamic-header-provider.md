# Dynamic Header Provider for OTLP HTTP Exporters

## Goal

Add the ability to inject dynamic HTTP headers into all OTLP HTTP exporters (traces, metrics, logs) via a custom `http.RoundTripper`. This allows callers of `SetupOTelSDK` to provide headers that are computed at request time (e.g., rotating auth tokens) rather than being fixed at initialization.

## Definitions

- **Header provider**: `func() map[string]string` called once per outbound HTTP request.
- **Dynamic transport**: `dynamicHeaderTransport` that wraps a base `http.RoundTripper` and applies provider headers before delegation.

## Requirements

### Functional
- `SetupOTelSDK` accepts functional options via a new `Option` type.
- A `WithHTTPClient(*http.Client)` option allows callers to provide a fully custom HTTP client to all three OTLP HTTP exporters.
- A `WithHeaderProvider(func() map[string]string)` option allows callers to provide a function that returns fresh headers on every HTTP request.
- When `WithHeaderProvider` is set, `SetupOTelSDK` constructs an `*http.Client` using a `dynamicHeaderTransport` round-tripper that calls the provider on each `RoundTrip`.
- When `WithHTTPClient` is set, the provided client is passed directly to all three exporters.
- If both are set, `WithHTTPClient` takes precedence (the caller is responsible for their own transport).
- The custom HTTP client is passed to each exporter via `otlptracehttp.WithHTTPClient`, `otlpmetrichttp.WithHTTPClient`, and `otlploghttp.WithHTTPClient`.
- `dynamicHeaderTransport` must clone the request (`req.Clone(req.Context())`) and clone headers before mutation to avoid mutating shared request state.
- Provider-returned headers are applied with `Header.Set(k, v)` and override any existing value for the same key.
- Provider results are validated defensively: skip empty header names; allow empty values.

### Non-Functional
- Zero behavior change when no options are provided — existing callers are unaffected.
- The `dynamicHeaderTransport` uses `http.DefaultTransport` as its base transport.
- Thread-safe: the header provider function may be called concurrently from multiple goroutines.
- Low allocation and minimal overhead per request (single request/header clone + map iteration).

## Scope

### In Scope
- New `Option` type and `setupConfig` struct in `pkg/telemetry`.
- `WithHTTPClient` and `WithHeaderProvider` option constructors.
- `dynamicHeaderTransport` implementing `http.RoundTripper`.
- Updated `SetupOTelSDK` signature: `func SetupOTelSDK(ctx context.Context, opts ...Option)`.
- Updated internal helper functions to accept and apply the HTTP client option.
- Tests for `dynamicHeaderTransport` and option application.
- Documentation comments describing precedence (`WithHTTPClient` over `WithHeaderProvider`) and per-request provider invocation behavior.

### Out of Scope
- Changes to any package outside `pkg/telemetry`.
- TLS configuration options.
- Retry or timeout configuration.
- Caching or memoization of provider output.

## Acceptance Criteria

1. `SetupOTelSDK(ctx)` (no options) compiles and behaves identically to today.
2. `SetupOTelSDK(ctx, WithHeaderProvider(fn))` injects headers from `fn` on every OTLP HTTP request.
3. `SetupOTelSDK(ctx, WithHTTPClient(client))` uses the provided client for all exporters.
4. `dynamicHeaderTransport` correctly delegates to the base transport after setting headers.
5. Unit tests cover:
   - transport header injection for traces/metrics/logs path usage,
   - option precedence (`WithHTTPClient` wins),
   - `nil` provider or `nil` returned map (no-op),
   - request/header cloning (no mutation of original request),
   - empty header key is ignored.
6. `make test` and `make check` pass.
