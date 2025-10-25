## MODIFIED Requirements
### Requirement: Shared Rule Registry
The system SHALL expose a canonical performance-by-default rule registry derived from the research methodology.
#### Scenario: Register new rule batch
- **WHEN** a tool loads the default registry
- **THEN** it SHALL expose rule records for `perf_avoid_linked_list`, `perf_large_enum_variant`, `perf_unnecessary_arc`, `perf_atomic_for_small_lock`, `perf_no_defer_in_loop`, `perf_avoid_rune_conversion`, `perf_needless_collect`, `perf_use_buffered_io`, and `perf_prefer_stack_alloc`, each with stable numeric identifiers.

#### Scenario: Provide rule guidance
- **WHEN** a language frontend queries any of the new rule records
- **THEN** the registry entry SHALL include the problem summary, detection hint, and fix guidance sourced from the "Proposed PerfCheck Rule Candidates" research so analyzers can emit actionable diagnostics without duplicating strings.
