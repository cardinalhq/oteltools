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

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/stretchr/testify/assert"
)

func Test_UrlScrub(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Scrub query string values",
			input:    "/path?query=value&page=10",
			expected: "/path?page=_&query=_",
		},
		{
			name:     "Scrub numbers in path",
			input:    "/tickets/591317/conversations",
			expected: "/tickets/<number>/conversations",
		},
		{
			name:     "Don't scrub camel case path segments",
			input:    "/api/v1/playMovie",
			expected: "/api/v1/playMovie",
		},
		{
			name:     "Scrub alphanumerics in path segments",
			input:    "/api/v1/playMovie/8UT64X",
			expected: "/api/v1/playMovie/<value>",
		},
		{
			name:     "Scrub uuids in path segments",
			input:    "/api/v1/playMovie/5960ff07-578a-4e49-a543-db92e8432860",
			expected: "/api/v1/playMovie/<value>",
		},
		{
			name:     "Don't scrub path parts that are not just numbers",
			input:    "/api/v1/graph",
			expected: "/api/v1/graph",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprFunc := urlScrub[any](&ottl.StandardStringGetter[any]{
				Getter: func(context.Context, any) (any, error) {
					return tt.input, nil
				},
			})
			result, err := exprFunc(context.Background(), nil)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
