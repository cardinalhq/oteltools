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

package authenv

import "errors"

type EnvironmentSource int

type EnvironmentSourceString string

const (
	EnvironmentSourceEnv EnvironmentSource = iota
	EnvironmentSourceAuth
)

var ErrInvalidEnvironmentSource = errors.New("invalid environment source, must be 'auth' or 'env'")

func ParseEnvironmentSource(s EnvironmentSourceString) (EnvironmentSource, error) {
	switch s {
	case "auth":
		return EnvironmentSourceAuth, nil
	case "env":
		return EnvironmentSourceEnv, nil
	default:
		return 0, ErrInvalidEnvironmentSource
	}
}
