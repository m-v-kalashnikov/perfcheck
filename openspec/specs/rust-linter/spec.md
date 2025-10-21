# rust-linter Specification

## Purpose
TBD - created by archiving change bootstrap-perfcheck. Update Purpose after archive.
## Requirements
### Requirement: Rust Performance Linter
The system SHALL provide a Rust linter capable of scanning crate source files for performance-by-default violations.

#### Scenario: Detect string concatenation in loops
- **WHEN** Rust source concatenates strings in a loop via `+=` or `format!`
- **THEN** the linter SHALL emit the associated rule warning.

#### Scenario: Detect missing vector preallocation
- **WHEN** Rust code grows a `Vec` within a counted loop without reserving capacity
- **THEN** the linter SHALL emit the associated rule warning.

#### Scenario: Detect dynamic dispatch in hot loops
- **WHEN** Rust code performs method calls through `dyn` trait objects inside a counted loop
- **THEN** the linter SHALL emit the `perf_avoid_reflection_dynamic` warning.

#### Scenario: Detect unbounded concurrency
- **WHEN** Rust code spawns threads or async tasks inside loops without using a limiter or bounded executor
- **THEN** the linter SHALL emit a `perf_bound_concurrency` diagnostic.

#### Scenario: Detect needless clones
- **WHEN** Rust code clones data inside a loop where borrowing would suffice
- **THEN** the linter SHALL emit a `perf_borrow_instead_of_clone` warning.

### Requirement: Linter CLI
The system SHALL expose a CLI entrypoint that analyzes a crate directory and reports violations with rule identifiers.

#### Scenario: CLI execution
- **WHEN** the CLI runs against a crate path
- **THEN** it SHALL traverse source files, apply all registered rules, and exit non-zero when violations are present.

