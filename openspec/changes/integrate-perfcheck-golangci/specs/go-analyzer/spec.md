## ADDED Requirements
### Requirement: GolangCI-Lint Native Integration
The system SHALL provide a GolangCI-Lint linter named `perfcheck` that executes the perfcheck analyzer suite when enabled through `.golangci.yml`, without requiring a separate vettool binary.

#### Scenario: Configuration-driven enablement
- **WHEN** a repository enables the `perfcheck` linter in GolangCI-Lint’s configuration
- **THEN** GolangCI-Lint SHALL execute the perfcheck analyzers during its run and report diagnostics under their existing rule IDs.

#### Scenario: Unsupported version handling
- **WHEN** GolangCI-Lint runs with a perfcheck module version that does not meet the documented compatibility matrix
- **THEN** the integration SHALL fail with a clear, actionable error that instructs the user how to resolve the version mismatch.
