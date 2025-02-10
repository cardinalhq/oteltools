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

import "context"

const (
	CardinalEnvCustomerID    = "CARDINALHQ_CUSTOMER_ID"
	CardinalEnvCollectorID   = "CARDINALHQ_COLLECTOR_ID"
	CardinalEnvCollectorName = "CARDINALHQ_COLLECTOR_NAME"
)

type Environment interface {
	CustomerID() string
	CollectorID() string
	CollectorName() string
	Tags() map[string]string
}

func GetEnvironment(ctx context.Context, source EnvironmentSource) Environment {
	if source == EnvironmentSourceAuth {
		return EnvironmentFromAuth(ctx)
	}
	return EnvironmentFromEnv()
}
