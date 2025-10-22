## ADDED Requirements
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
