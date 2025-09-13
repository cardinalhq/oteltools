# AGENTS.md

## Development Commands

### Build and Test

```bash
# Generate code from protobuf and Ragel files
make generate

# Run tests with race detection
make test

# Run benchmarks
make bench

# Run pre-commit checks (includes license headers and linting)
make check
```

### Code Quality

```bash
# Check license headers
go tool license-eye header check

# Run golangci-lint
go tool golangci-lint run
```

### Single Test Commands

```bash
# Run a specific test
go test -race ./pkg/ottl -run TestSpecificFunction

# Run tests in a specific package
go test -race ./signalbuilder/...

# Run benchmarks for a specific package
go test -bench=. ./pkg/ottl/...
```

## Architecture Overview

This is a shared Go library (`github.com/cardinalhq/oteltools`) providing common tools for OpenTelemetry collector components and backend services used by CardinalHQ.

### Key Packages

**signalbuilder/** - Core package for efficiently building OpenTelemetry signals

- Provides builders for metrics, traces, and logs with automatic deduplication
- Optimizes memory through resource/scope caching and hash-based lookups
- Follows OTEL hierarchy: Resource → Scope → Signal
- Supports both protobuf (pdata) and structured (JSON/YAML) output

**pkg/ottl/** - OpenTelemetry Transformation Language extensions

- Custom OTTL functions for data transformation and enrichment
- Sampling strategies (RPS-based, static)
- Aggregation and filtering capabilities
- Configuration management for transformation rules

**pkg/chqpb/** - CardinalHQ Protocol Buffers

- Custom protobuf definitions for internal data structures
- Statistics caching and event tracking
- Hash-based key generation for efficient data organization

**pkg/authenv/** - Authentication Environment

- Abstracts collector authentication across different deployment models
- Supports both environment variables (customer sites) and client auth (SaaS)
- Provides collector ID, customer ID, and environment context

**pkg/fingerprinter/** - Data fingerprinting and tokenization

- Tokenizes data for efficient processing and deduplication
- Uses Ragel state machines for parsing

**pkg/filereader/** - File reading abstractions

- Supports both local and HTTP-based file access
- Used for configuration and data file loading

**hashutils/**, **maputils/**, **stringutils/** - Utility packages

- Common hash functions, map operations, and string processing
- Anti-ANSI string cleaning and quoted string parsing

### Protocol Buffers

The codebase uses protobuf extensively for data serialization. Generated Go files are created via `make generate`

### Code Generation

Uses Ragel for state machine generation in tokenizer packages:

- `pkg/fingerprinter/tokenizer/`
- `pkg/pii/tokenizer/`

### Dependencies

- Built on Go 1.24
- Uses OpenTelemetry Collector and contrib packages extensively
- Leverages DataDog sketches and Apache DataSketches for statistical computations
- Includes golangci-lint and license-eye as development tools

### Testing Patterns

- Race detection enabled by default (`go test -race`)
- Extensive use of testdata directories for test fixtures
- Mock generation for interfaces
- Benchmark tests for performance-critical code paths

### License Management

All source files must include Apache 2.0 license headers. Use `make check` to verify compliance.
