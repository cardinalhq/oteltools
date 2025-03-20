# Copyright 2024-2025 CardinalHQ, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Generate all the things.
#
generate:
	go generate ./...

#
# Run pre-commit checks
#
check: test
	go tool license-eye header check
	go tool golangci-lint run

#
# Test targets
#

.PHONY: test
test: generate
	go test -race ./...

.PHONY: bench benchmark
bench benchmark: generate
	go test -bench=. ./...

#
# Clean the world.
#

.PHONY: clean
clean:

.PHONY: really-clean
really-clean: clean
	rm -f ${pb_deps}
