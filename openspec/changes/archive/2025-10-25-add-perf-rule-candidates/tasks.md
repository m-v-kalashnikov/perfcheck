## 1. Specification
- [x] 1.1 Finalize proposal and spec deltas for the nine new PerfCheck rules.

## 2. Implementation
- [x] 2.1 Add the nine rule definitions to `perfcheck-core/config/default_rules.tsv` and update the schema fixtures if needed.
- [x] 2.2 Implement Go analyzer detectors for `perf_avoid_linked_list`, `perf_atomic_for_small_lock`, `perf_no_defer_in_loop`, `perf_avoid_rune_conversion`, `perf_use_buffered_io`, and `perf_prefer_stack_alloc`, each with unit tests.
- [x] 2.3 Implement Rust linter detectors for `perf_avoid_linked_list`, `perf_large_enum_variant`, `perf_unnecessary_arc`, `perf_atomic_for_small_lock`, `perf_needless_collect`, and `perf_prefer_stack_alloc`, each with unit tests.
- [x] 2.4 Ensure both analyzers surface fix guidance sourced from the updated registry entries.
- [x] 2.5 Update user-facing docs or release notes to describe the new rules and any migration guidance.

## 3. Validation
- [x] 3.1 Run `openspec validate add-perf-rule-candidates --strict`.
- [x] 3.2 Run `go test ./...` under `go/` and `cargo test` under `rust/` after adding the detectors.
