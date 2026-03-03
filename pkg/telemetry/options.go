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

import "net/http"

// Option configures SetupOTelSDK behavior.
type Option func(*setupConfig)

type setupConfig struct {
	httpClient     *http.Client
	headerProvider func() map[string]string
}

// WithHTTPClient sets a custom HTTP client for all OTLP exporters.
// Takes precedence over WithHeaderProvider.
func WithHTTPClient(client *http.Client) Option {
	return func(c *setupConfig) {
		c.httpClient = client
	}
}

// WithHeaderProvider sets a function that returns headers to inject
// into every outbound OTLP HTTP request.
func WithHeaderProvider(provider func() map[string]string) Option {
	return func(c *setupConfig) {
		c.headerProvider = provider
	}
}

type dynamicHeaderTransport struct {
	base     http.RoundTripper
	provider func() map[string]string
}

func (t *dynamicHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	if t.provider == nil {
		return base.RoundTrip(req)
	}

	headers := t.provider()
	if len(headers) == 0 {
		return base.RoundTrip(req)
	}

	req = req.Clone(req.Context())
	for k, v := range headers {
		if k == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	return base.RoundTrip(req)
}

// resolveHTTPClient returns the *http.Client to use, or nil if defaults should apply.
func (cfg *setupConfig) resolveHTTPClient() *http.Client {
	if cfg == nil {
		return nil
	}
	if cfg.httpClient != nil {
		return cfg.httpClient
	}
	if cfg.headerProvider != nil {
		return &http.Client{
			Transport: &dynamicHeaderTransport{
				base:     http.DefaultTransport,
				provider: cfg.headerProvider,
			},
		}
	}
	return nil
}
