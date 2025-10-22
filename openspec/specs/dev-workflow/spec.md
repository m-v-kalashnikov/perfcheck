# dev-workflow Specification

## Purpose
TBD - created by archiving change expand-rule-coverage. Update Purpose after archive.
## Requirements
### Requirement: Repository Smoke Test
The system SHALL expose a single command that runs Go unit tests, Rust unit tests, and `openspec validate --strict`.

#### Scenario: CLI smoke run
- **WHEN** a developer executes the documented smoke-test command
- **THEN** it SHALL run `go test ./...`, `cargo test`, and `openspec validate --strict`, failing fast on the first error.

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
