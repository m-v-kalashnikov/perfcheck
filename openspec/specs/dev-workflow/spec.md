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

### Requirement: Rust Tooling Configuration Baseline
The developer workflow SHALL version clippy, rustfmt, and Taplo configuration files inside the `rust/` workspace directory so that local and CI runs share identical lint and formatting behavior.

#### Scenario: Clippy honors repository configuration
- **WHEN** a contributor runs `cargo clippy` through the documented maintenance command
- **THEN** Clippy SHALL apply the rules specified in `rust/clippy.toml`, including any enforced deny lists.

#### Scenario: Rustfmt honors repository configuration
- **WHEN** a contributor runs `cargo fmt --check` via the maintenance workflow
- **THEN** rustfmt SHALL read `rust/rustfmt.toml` to determine formatting style, producing deterministic output across toolchain versions.

#### Scenario: Taplo honors repository configuration
- **WHEN** a contributor runs the documented `just rust-maintain` workflow
- **THEN** the Taplo step SHALL execute using `rust/taplo.toml` so configuration files and fixtures stay consistently formatted.

### Requirement: Go Tooling Configuration Baseline
The developer workflow SHALL version Go formatting and lint configuration inside the `go/` directory so that import formatting and lint behavior remain stable across contributors and CI.

#### Scenario: Goimports honors repository configuration
- **WHEN** a contributor formats Go files using the documented goimports command
- **THEN** goimports SHALL apply the repository-local import grouping (for example by invoking `goimports -local github.com/m-v-kalashnikov/perfcheck`) so local modules remain grouped consistently.

#### Scenario: GolangCI-Lint honors repository configuration
- **WHEN** a contributor runs the documented Go lint workflow
- **THEN** GolangCI-Lint SHALL source its configuration from the checked-in `go/.golangci.yml`, applying the enforced performance rule set and any documented exceptions.

