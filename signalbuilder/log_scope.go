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

package signalbuilder

import (
	"go.opentelemetry.io/collector/pdata/plog"
)

type LogScopeBuilder struct {
	scope plog.ScopeLogs
}

func NewLogScopeBuilder(scope plog.ScopeLogs) *LogScopeBuilder {
	return &LogScopeBuilder{
		scope: scope,
	}
}

func (lsb *LogScopeBuilder) AddRecord() plog.LogRecord {
	return lsb.scope.LogRecords().AppendEmpty()
}

func (lsb *LogScopeBuilder) Get() plog.ScopeLogs {
	return lsb.scope
}
