## Why
The shared rule registry already enumerates the performance-by-default guidance, but both analyzers cover only a subset. Completing rule coverage ensures Go and Rust developers receive consistent feedback for every documented rule.

## What Changes
- Implement Go analyzers for the remaining registry rules (reflection in loops, bounded concurrency, EqualFold comparisons, sync.Pool pointer storage, byte-oriented writers).
- Extend the Rust linter to cover outstanding registry rules (bounded concurrency, avoid needless clones, reflection/dynamic dispatch parity gaps).
- Expand documentation and test fixtures to keep examples and smoke tests aligned with the full rule set.

## Impact
- Affected specs: go-analyzer, rust-linter, documentation
- Affected code: `go/internal/analyzer`, `go/internal/ruleset`, `rust/src/linter.rs`, `rust/src/rules.rs`, test suites, docs
