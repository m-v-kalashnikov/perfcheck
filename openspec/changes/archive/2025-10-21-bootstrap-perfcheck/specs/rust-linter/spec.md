## ADDED Requirements
### Requirement: Rust Performance Linter
The system SHALL provide a Rust linter capable of scanning crate source files for performance-by-default violations.

#### Scenario: Detect string concatenation in loops
- **WHEN** Rust source concatenates strings in a loop via `+=` or `format!`
- **THEN** the linter SHALL emit the associated rule warning.

#### Scenario: Detect missing vector preallocation
- **WHEN** Rust code grows a `Vec` within a counted loop without reserving capacity
- **THEN** the linter SHALL emit the associated rule warning.

### Requirement: Linter CLI
The system SHALL expose a CLI entrypoint that analyzes a crate directory and reports violations with rule identifiers.

#### Scenario: CLI execution
- **WHEN** the CLI runs against a crate path
- **THEN** it SHALL traverse source files, apply all registered rules, and exit non-zero when violations are present.
