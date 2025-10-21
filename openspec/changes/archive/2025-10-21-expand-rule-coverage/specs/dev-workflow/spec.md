## ADDED Requirements
### Requirement: Repository Smoke Test
The system SHALL expose a single command that runs Go unit tests, Rust unit tests, and `openspec validate --strict`.

#### Scenario: CLI smoke run
- **WHEN** a developer executes the documented smoke-test command
- **THEN** it SHALL run `go test ./...`, `cargo test`, and `openspec validate --strict`, failing fast on the first error.
