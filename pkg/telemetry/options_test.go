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
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestDynamicHeaderTransport_InjectsHeaders(t *testing.T) {
	var captured http.Header
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req.Header
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base: base,
		provider: func() map[string]string {
			return map[string]string{
				"Authorization": "Bearer token123",
				"X-Custom":      "value",
			}
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "Bearer token123", captured.Get("Authorization"))
	assert.Equal(t, "value", captured.Get("X-Custom"))
}

func TestDynamicHeaderTransport_ClonesRequest(t *testing.T) {
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base: base,
		provider: func() map[string]string {
			return map[string]string{"X-Injected": "yes"}
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Original", "keep")
	originalHeaders := req.Header.Clone()

	_, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, originalHeaders, req.Header, "original request headers must not be mutated")
}

func TestDynamicHeaderTransport_NilProvider(t *testing.T) {
	called := false
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base:     base,
		provider: nil,
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDynamicHeaderTransport_NilMap(t *testing.T) {
	var captured *http.Request
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base: base,
		provider: func() map[string]string {
			return nil
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Same(t, req, captured, "should pass original request when map is nil")
}

func TestDynamicHeaderTransport_EmptyKeyIgnored(t *testing.T) {
	var captured http.Header
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req.Header
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base: base,
		provider: func() map[string]string {
			return map[string]string{
				"":       "should-be-skipped",
				"X-Good": "kept",
			}
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, "kept", captured.Get("X-Good"))
	assert.Empty(t, captured.Get(""))
}

func TestDynamicHeaderTransport_OverrideExistingHeader(t *testing.T) {
	var captured http.Header
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req.Header
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &dynamicHeaderTransport{
		base: base,
		provider: func() map[string]string {
			return map[string]string{"Authorization": "new-token"}
		},
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Authorization", "old-token")
	_, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, "new-token", captured.Get("Authorization"))
}

func TestDynamicHeaderTransport_NilBase(t *testing.T) {
	// With nil base, should fall back to http.DefaultTransport (no panic).
	transport := &dynamicHeaderTransport{
		base:     nil,
		provider: func() map[string]string { return map[string]string{"X-Test": "val"} },
	}

	var captured http.Header
	// Replace DefaultTransport temporarily to avoid real network calls.
	orig := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req.Header
		return &http.Response{StatusCode: 200}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = orig })

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "val", captured.Get("X-Test"))
}

func TestDynamicHeaderTransport_ConcurrentAccess(t *testing.T) {
	transport := &dynamicHeaderTransport{
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200}, nil
		}),
		provider: func() map[string]string {
			return map[string]string{"X-Auth": "token"}
		},
	}

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			req.Header.Set("X-Goroutine", fmt.Sprintf("%d", i))
			resp, err := transport.RoundTrip(req)
			assert.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
	wg.Wait()
}

func TestResolveHTTPClient_Precedence(t *testing.T) {
	explicit := &http.Client{}
	cfg := &setupConfig{
		httpClient: explicit,
		headerProvider: func() map[string]string {
			return map[string]string{"X-Ignored": "true"}
		},
	}
	result := cfg.resolveHTTPClient()
	assert.Same(t, explicit, result, "WithHTTPClient must take precedence")
}

func TestResolveHTTPClient_HeaderProviderOnly(t *testing.T) {
	cfg := &setupConfig{
		headerProvider: func() map[string]string {
			return map[string]string{"X-Auth": "token"}
		},
	}
	result := cfg.resolveHTTPClient()
	require.NotNil(t, result)
	_, ok := result.Transport.(*dynamicHeaderTransport)
	assert.True(t, ok, "should wrap with dynamicHeaderTransport")
}

func TestResolveHTTPClient_ZeroConfig(t *testing.T) {
	cfg := &setupConfig{}
	assert.Nil(t, cfg.resolveHTTPClient())
}

func TestResolveHTTPClient_NilReceiver(t *testing.T) {
	var cfg *setupConfig
	assert.Nil(t, cfg.resolveHTTPClient())
}
