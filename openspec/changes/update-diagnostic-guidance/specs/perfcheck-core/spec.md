## MODIFIED Requirements
### Requirement: Shared Rule Registry
The system SHALL expose a canonical performance-by-default rule registry derived from the research methodology.

#### Scenario: Load default rules
- **WHEN** tools request the default rule set
- **THEN** the registry SHALL return structured rule metadata validated against the schema.

#### Scenario: Numeric identifiers
- **WHEN** a tool loads the registry
- **THEN** each rule SHALL provide a deterministic numeric code for hot-path lookups.

#### Scenario: Provide guidance metadata
- **WHEN** a tool reads a rule definition
- **THEN** the registry SHALL supply human-readable `problem_summary` and `fix_hint` fields that analyzers can embed directly into diagnostics.

### Requirement: Rule Schema Validation
The system SHALL provide a machine-readable schema so language analyzers can validate rule definitions.

#### Scenario: Schema availability
- **WHEN** a tool bundles the registry
- **THEN** it SHALL embed the schema and default rules for offline use.
