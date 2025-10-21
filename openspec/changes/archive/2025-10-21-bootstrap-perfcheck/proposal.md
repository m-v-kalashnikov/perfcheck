## Why
perfcheck needs an initial baseline that turns the Performance-by-Default research into enforceable tooling.

## What Changes
- Bootstrap a shared rule registry derived from the methodology research.
- Provide a Go analyzer binary that surfaces the first set of performance-by-default checks.
- Provide a Rust linter crate capable of running the shared rules against Rust sources.
- Wire in documentation and defaults so both analyzers stay in sync with the rule schema.

## Impact
- Affected specs: perfcheck-core (new capability), go-analyzer (new capability), rust-linter (new capability)
- Affected code: perfcheck-core/, go/, rust/
