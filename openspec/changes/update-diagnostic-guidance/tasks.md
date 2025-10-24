## 1. Registry Enhancements
- [x] 1.1 Extend the rule schema and TSV to include `problem_summary` and `fix_hint` fields populated for existing rules.
- [x] 1.2 Add validation/tests to prevent empty guidance fields when new rules are registered.

## 2. Analyzer Diagnostics
- [x] 2.1 Update Go analyzer diagnostics (CLI + go/analysis) to display the explanation and fix suggestion, with golden tests covering formatting.
- [x] 2.2 Update the Rust CLI output and snapshot tests to surface the same guidance content for every rule.

## 3. Documentation & Tooling
- [x] 3.1 Document the diagnostic message format in `docs/performance-by-default.md` and update examples to match the richer guidance.
- [x] 3.2 Ensure developer workflows (lint/test scripts) exercise the new validation so regressions are caught in CI.
