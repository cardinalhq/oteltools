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

package functions

import (
	"context"
	"github.com/oschwald/geoip2-golang"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/stretchr/testify/assert"
)

// City is the mock implementation of geoip2.Reader's City method

func testFile(file string) string {
	return filepath.Join("metadata", file)
}

func Test_IpLocation_ValidIP(t *testing.T) {
	db, err := geoip2.Open(initDb()) // Open the GeoLite2 database

	exprFunc := iplocation[any](db, &ottl.StandardStringGetter[any]{
		Getter: func(context.Context, any) (any, error) {
			return "73.202.180.160", nil
		},
	})

	result, err := exprFunc(context.Background(), nil)
	assert.NoError(t, err)

	expected := map[string]any{
		"city":      "Danville",
		"country":   "United States",
		"zip_code":  "94506",
		"latitude":  37.8333,
		"longitude": -121.9209,
	}

	assert.Equal(t, expected, result)
}
