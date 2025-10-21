## 1. Go Analyzer
- [x] 1.1 Add loop reflection detection for `perf_avoid_reflection_dynamic` and unit tests.
- [x] 1.2 Enforce `perf_bound_concurrency` for Go concurrency hotspots with tests.
- [x] 1.3 Implement `perf_equal_fold_compare` detection in Go code and tests.
- [x] 1.4 Add sync.Pool pointer storage rule (`perf_syncpool_store_pointers`) with coverage.
- [x] 1.5 Report writer byte-preference (`perf_writer_prefer_bytes`) when string conversions occur.

## 2. Rust Linter
- [x] 2.1 Extend concurrency tracking for `perf_bound_concurrency` detection.
- [x] 2.2 Flag unnecessary cloning with `perf_borrow_instead_of_clone`.
- [x] 2.3 Ensure dynamic-dispatch coverage matches registry expectations (map to `perf_avoid_reflection_dynamic`).

## 3. Docs & Tooling
- [x] 3.1 Add documentation examples for the newly covered rules.
- [x] 3.2 Update smoke tests or fixtures if additional sample projects are required.
- [x] 3.3 Replace `scripts/smoke.sh` with a `just` recipe that runs Go tests, Rust tests, and OpenSpec validation.
- [x] 3.4 Build Go and Rust fixture packages/crates that intentionally trigger each rule for regression testing.
- [x] 3.5 Expand README or docs with a rule matrix linking to the new examples and fixtures.
