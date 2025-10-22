# dev-workflow Specification

## Purpose
TBD - created by archiving change expand-rule-coverage. Update Purpose after archive.
## Requirements
### Requirement: Repository Pre-Commit Run
The system SHALL expose a single command that runs Go unit tests, Rust unit tests, and `openspec validate --strict`.

#### Scenario: CLI pre-commit run
- **WHEN** a developer executes the documented pre-commit command
- **THEN** it SHALL run `go test ./...`, `cargo nextest run`, and `openspec validate --strict`, failing fast on the first error.

### Requirement: Tooling Entry Point
The developer workflow SHALL document the `Justfile` as the single supported entry point for running project automation.

#### Scenario: Running workflows
- **WHEN** a developer needs to execute linting, auditing, or test workflows
- **THEN** they SHALL invoke the corresponding `just` recipes rather than calling underlying tools directly.

### Requirement: Language Scoped Tooling Configuration
The repository SHALL keep language-specific configuration and helper binaries alongside their respective language directories while delegating orchestration to the root `Justfile`.

#### Scenario: Tool configuration layout
- **WHEN** language tooling (e.g., GolangCI-Lint, cargo-deny) requires configuration files or wrapper binaries
- **THEN** those assets SHALL reside within the corresponding language directory (`go/`, `rust/`, …), and the root `Justfile` SHALL remain the canonical interface for invoking them.

### Requirement: Go Lint Command
The developer workflow SHALL expose a documented command that executes `golangci-lint run` against the repository using a checked-in configuration and fails when diagnostics are reported.

#### Scenario: Go lint execution
- **WHEN** a developer runs the documented Go lint command from the repository root
- **THEN** it SHALL invoke `golangci-lint run` with the repository configuration file and exit non-zero if any lints fail.

### Requirement: Rust Maintenance Command
The developer workflow SHALL expose a documented command that sequentially runs `cargo fmt --check`, `cargo clippy --all-targets --all-features -D warnings`, `cargo deny check`, `cargo audit`, and `cargo udeps --all-targets`, stopping on the first failure.

#### Scenario: Rust toolchain execution
- **WHEN** a developer runs the documented Rust maintenance command
- **THEN** it SHALL execute each configured cargo tool in order, aborting on the first failing step, and exit non-zero when any tool reports an issue.

