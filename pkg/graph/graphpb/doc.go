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

// Package graphpb holds source and generated code for the graph objects that we
// collect from Kubernetes and other sources.  Each top-level "summary" object
// is a description of a Kubernetes resource definition, and the PackagedObject
// includes the additional OpenTelemetry resource and log record attributes.
package graphpb

//go:generate sh -c "protoc -I . --go_out=. --go_opt=paths=source_relative *.proto"
//go:generate sh -c "rm -f baseobject_getter.go ; go run tools/gen_baseobject_getter.go -o baseobject_getter.go"
