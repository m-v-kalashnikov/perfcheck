## 1. Rule Coverage
- [x] 1.1 Add Go analyzer detection for `perf_preallocate_collections`, including unit tests and rule registration.
- [x] 1.2 Extend the Rust linter with an additional rule (`perf_avoid_reflection_dynamic`), covering detection logic and tests.

## 2. Validation Workflow
- [x] 2.1 Introduce a smoke-test command/script that runs Go tests, Rust tests, and `openspec validate --strict`.

## 3. Documentation
- [x] 3.1 Update `docs/performance-by-default.md` with examples for the new Go and Rust rules and reference the smoke test entrypoint.
