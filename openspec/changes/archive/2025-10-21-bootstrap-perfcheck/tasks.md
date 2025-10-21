## 1. Shared Rule Registry
- [x] 1.1 Define rule schema and embed default rules for reuse.
- [x] 1.2 Provide loader helpers in Go and Rust that expose numeric rule identifiers.

## 2. Go Analyzer
- [x] 2.1 Implement go/analysis-based lints for hot string concatenation and regex compilation.
- [x] 2.2 Expose analyzers via a `unitchecker` main and add tests.

## 3. Rust Linter
- [x] 3.1 Implement AST visitor detecting string concatenation in loops and vector preallocation gaps.
- [x] 3.2 Provide CLI entrypoint and unit tests for the Rust lint engine.

## 4. Documentation & Defaults
- [x] 4.1 Sync default_rules.tsv with methodology research and document usage.
- [x] 4.2 Add README updates covering analyzer usage and testing instructions.
