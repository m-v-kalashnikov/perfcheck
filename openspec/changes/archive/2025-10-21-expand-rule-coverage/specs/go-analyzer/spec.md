## MODIFIED Requirements
### Requirement: Go Performance Analyzer
The system SHALL provide a go/analysis-based linter that surfaces performance-by-default violations.

#### Scenario: Detect string concatenation in loops
- **WHEN** a Go package contains string concatenation inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

#### Scenario: Detect regex compilation inside loops
- **WHEN** a Go package compiles a regexp inside a loop
- **THEN** the analyzer SHALL report a warning referencing the corresponding rule id.

#### Scenario: Detect missing collection preallocation
- **WHEN** a Go package appends to a slice or map within a deterministic loop without reserving capacity when the final size is known
- **THEN** the analyzer SHALL report a warning referencing `perf_preallocate_collections`.
