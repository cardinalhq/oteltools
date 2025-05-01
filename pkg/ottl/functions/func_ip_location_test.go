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
	"testing"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IpLocation_ValidIP(t *testing.T) {
	db, err := maxminddb.Open("../testdata/GeoIP2-Country-Test.mmdb")
	require.NoError(t, err)
	defer db.Close()

	exprFunc := iplocation(db, &ottl.StandardStringGetter[any]{
		Getter: func(context.Context, any) (any, error) {
			return "2a02:d300::1", nil
		},
	})

	result, err := exprFunc(context.Background(), nil)
	require.NoError(t, err)

	expected := map[string]any{
		"city":        "Unknown",
		"country":     "Ukraine",
		"zip_code":    "",
		"country_iso": "UA",
		"latitude":    0.0,
		"longitude":   0.0,
	}

	assert.Equal(t, expected, result)
}
