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

// Environment is an interface that provides access to common items
// like the collector ID that is running, and the customer ID running
// the collector.
//
// Two sources of environment information are supported:
// 1. The environment variables set in the process.
// 2. The attributes set in the client.Info.Auth object.
//
// The environment variables are used when the collector is running as a
// regular collector in a customer site.  The from-auth version is used
// when the collector is running as a component of the CardinalHQ
// SaaS collector.
//
// Each collector's auth configuration, when posting to the SaaS collector
// or other components like external-api, will always have a customer ID
// and collector ID.  Additionally, we can get some other data items such
// as the collector name and other environment variables set in the collector
// config.
