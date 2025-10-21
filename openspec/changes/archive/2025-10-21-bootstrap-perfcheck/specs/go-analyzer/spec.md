## ADDED Requirements
### Requirement: Go Performance Analyzer
The system SHALL provide a go/analysis-based linter that surfaces performance-by-default violations.

#### Scenario: Detect string concatenation in loops
- **WHEN** a Go package contains string concatenation inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

#### Scenario: Detect regex compilation inside loops
- **WHEN** a Go package compiles a regexp inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

### Requirement: Analyzer Packaging
The system SHALL expose the analyzer as a unitchecker-compatible binary for integration with go vet and golangci-lint.

#### Scenario: CLI entrypoint
- **WHEN** `go vet -vettool` is invoked with the perfcheck analyzer
- **THEN** it SHALL execute the registered performance checks without additional setup.
