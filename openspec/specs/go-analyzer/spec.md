# go-analyzer Specification

## Purpose
TBD - created by archiving change bootstrap-perfcheck. Update Purpose after archive.
## Requirements
### Requirement: Go Performance Analyzer
The system SHALL provide a go/analysis-based linter that surfaces performance-by-default violations.

#### Scenario: Detect string concatenation in loops
- **WHEN** a Go package contains string concatenation inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

#### Scenario: Detect regex compilation inside loops
- **WHEN** a Go package compiles a regexp inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

#### Scenario: Detect missing collection preallocation
- **WHEN** a Go package appends to a slice or map within a deterministic loop without reserving capacity when the final size is known
- **THEN** the analyzer SHALL report a warning referencing `perf_preallocate_collections`.

#### Scenario: Detect reflection usage inside hot loops
- **WHEN** Go code performs `reflect` operations or type assertions on every iteration of a loop
- **THEN** the analyzer SHALL report a `perf_avoid_reflection_dynamic` warning.

#### Scenario: Detect unbounded goroutine spawning
- **WHEN** Go code launches new goroutines inside loops without using a worker pool or semaphore limit
- **THEN** the analyzer SHALL report a `perf_bound_concurrency` diagnostic.

#### Scenario: Detect case-insensitive comparisons using ToLower/ToUpper
- **WHEN** Go code normalizes strings with `strings.ToLower`/`strings.ToUpper` for equality checks
- **THEN** the analyzer SHALL report a `perf_equal_fold_compare` diagnostic recommending `strings.EqualFold`.

#### Scenario: Detect sync.Pool value storage without pointers
- **WHEN** Go code stores non-pointer values in a `sync.Pool`
- **THEN** the analyzer SHALL emit a `perf_syncpool_store_pointers` warning.

#### Scenario: Detect writer string conversions
- **WHEN** Go code converts byte slices to strings solely for writing to an `io.Writer`
- **THEN** the analyzer SHALL emit a `perf_writer_prefer_bytes` warning.

### Requirement: Analyzer Packaging
The system SHALL expose the analyzer as a unitchecker-compatible binary for integration with go vet and golangci-lint.

#### Scenario: CLI entrypoint
- **WHEN** `go vet -vettool` is invoked with the perfcheck analyzer
- **THEN** it SHALL execute the registered performance checks without additional setup.

