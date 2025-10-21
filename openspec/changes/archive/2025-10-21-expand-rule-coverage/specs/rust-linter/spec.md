## MODIFIED Requirements
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
