# documentation Specification

## Purpose
TBD - created by archiving change expand-rule-coverage. Update Purpose after archive.
## Requirements
### Requirement: Performance Rule Examples
The documentation SHALL include illustrative code snippets for each enforced analyzer rule.

#### Scenario: Document new Go and Rust rules
- **WHEN** the project adds a new Go or Rust analyzer rule
- **THEN** `docs/performance-by-default.md` SHALL gain an example explaining the violation and preferred fix.

#### Scenario: Maintain complete rule coverage examples
- **WHEN** all registry rules are enforced by the analyzers
- **THEN** the documentation SHALL contain at least one example per rule so contributors understand each diagnostic.

