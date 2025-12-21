# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests with race detection and coverage
go test -v -race -cover ./...

# Run a single test
go test -v -run TestFaultSetEnabled ./...

# Run benchmarks
go test -run=XXX -bench=.

# Run linter (used in CI)
golangci-lint run

# Check formatting
gofmt -l -w -s ./
```

## CI Requirements

- Tests must pass with race detection enabled
- Code coverage must be exactly 100%
- golangci-lint must pass (see `.golangci.yml` for enabled linters)

## Architecture

This is an HTTP fault injection library for Go. The core abstraction is:

**Fault** (`fault.go`) - Wrapper that controls when an Injector runs. Handles:
- Enable/disable state
- Participation percentage (what % of requests trigger the fault)
- Path and header allowlists/blocklists
- Random seed for reproducibility

**Injector** (`injector.go`) - Interface for fault implementations. Each injector implements `Handler(next http.Handler) http.Handler`:
- `ErrorInjector` - Returns HTTP error status codes
- `SlowInjector` - Adds latency via configurable sleep function
- `RejectInjector` - Returns empty response using `panic(http.ErrAbortHandler)`
- `ChainInjector` - Runs multiple injectors sequentially
- `RandomInjector` - Randomly picks one injector to run

**Reporter** (`reporter.go`) - Interface for observability. Receives `StateStarted`/`StateFinished`/`StateSkipped` events.

## Code Patterns

- Options pattern: Each type has its own `*Option` interface with `apply*` methods
- Constructors validate options and return errors
- Thread-safe: `rand.Rand` protected by mutex in `Fault` and `RandomInjector`
- All randomness seeded with `defaultRandSeed(1)` by default for reproducibility
