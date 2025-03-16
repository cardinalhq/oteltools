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

# Due to the way we build, we will make the universe no matter which files
# actually change.  With the many targets, this is just so much easier,
# and it also ensures the Docker images have identical timestamp-based tags.
pb_deps = pkg/chqpb/attribute.pb.go \
	pkg/chqpb/kafkaenvelope.pb.go \
	pkg/chqpb/log_ingest_stats.pb.go \
	pkg/chqpb/metric_ingest_stats.pb.go \
	pkg/chqpb/span_ingest_stats.pb.go \
	pkg/chqpb/stats.pb.go \

#
# Generate all the things.
#
generate: ${pb_deps}

#
# Run pre-commit checks
#
check: test
	go tool license-eye header check
	go tool golangci-lint run

#
# build protobufs
#
${pb_deps}: pkg/chqpb/%.pb.go: pkg/chqpb/%.proto
	go generate ./...

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
