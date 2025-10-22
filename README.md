# perfcheck

Cross-language **Performance-by-Default** linting toolkit for Go and Rust. The goal is to codify the research in `docs/` into analyzers that flag inefficient patterns before they reach production.

## Structure
- `perfcheck-core/` — canonical rule registry (`config/default_rules.tsv`) and schema docs
- `go/` — Go analyzer executable (unitchecker) plus integration tests
- `rust/` — Rust linter library + CLI that walks crates
- `docs/` — methodology and research source material
- `openspec/` — change tracking and capability specs

## Shared Rule Registry
- Authoritative TSV lives at `perfcheck-core/config/default_rules.tsv`; schema documented in `perfcheck-core/schema/rule_schema.json`.
- Go embeds a copy via `go/internal/ruleset`. Run `go generate ./internal/ruleset` inside `go/` after editing the TSV to refresh the embedded bundle.
- Rust loads the TSV directly at runtime through `include_str!` to avoid drift.

## Go Analyzer
- Build: `cd go && go build ./cmd/perfcheck-go`
- Run (example): `go vet -vettool=$(pwd)/perfcheck-go ./...`
- GolangCI-Lint integration: build `go/cmd/perfcheck-golangci` and register it as a custom linter (see `docs/integrations.md#golangci-lint`).
- Tests: `cd go && GOCACHE=$(pwd)/.gocache go test ./...`

## Rust Linter
- Build/Run: `cd rust && cargo run --bin perfcheck -- <crate-path>`
- Tests: `cd rust && cargo nextest run`
- Clippy integration: install the workspace binaries and run `cargo perfcheck-clippy` (details in `docs/integrations.md#clippy`).
- Exports `lint_source`/`lint_path` helpers for embedding into other tooling.
- Emits diagnostics as `path:line:column [rule] message` and exits non-zero if any violations are found.

## Development Notes
- Keep rule IDs stable; they double as numeric hashes for hot-path lookups.
- Update docs/performance-by-default.md when expanding the rule set.
- Install local tooling before linting: `golangci-lint`, `goimports`, `taplo` (`cargo install taplo-cli`), `rustup component add rustfmt clippy`, `rustup toolchain install nightly`, and `cargo install cargo-deny cargo-audit cargo-udeps` (these pull advisory databases on first run).
- Run `just go-maintain` to apply `golangci-lint fmt` (running `gofmt`, `goimports`, `gci`, and `golines` with the configured rewrites), build the GolangCI-Lint bridge, run `golangci-lint run`, execute the perfcheck multichecker, verify `go.mod`, and scan with `govulncheck` (first run downloads the Go vulnerability database).
- Run `just rust-maintain` to apply `cargo fmt --check`, validate TOML manifests with `taplo format --check --diff`, execute `cargo clippy --all-targets --all-features -- -D warnings`, and run the deny/audit/udeps hygiene checks (deny/audit steps require the RustSec database, so ensure network access when refreshing it).
- Run `just pre-commit` from the repository root to execute the lint commands, Go tests, Rust tests, and `openspec validate --strict` before submitting changes.

## Rule Matrix

| Rule ID                         | Languages | Description                                                                              | Docs                                                                     | Fixtures                                                                                              |
|---------------------------------|-----------|------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| `perf_avoid_string_concat_loop` | Go, Rust  | Avoid string concatenation in loops; use builders or reserved buffers                    | [Docs](docs/performance-by-default.md#perf_avoid_string_concat_loop)     | Go: `go/internal/analyzer/testdata/src/violations/violations.go`, Rust: `rust/fixtures/violations.rs` |
| `perf_regex_compile_once`       | Go        | Compile regular expressions once instead of inside hot loops                             | [Docs](docs/performance-by-default.md#perf_regex_compile_once-go)        | Go: `go/internal/analyzer/testdata/src/violations/violations.go`                                      |
| `perf_preallocate_collections`  | Go, Rust  | Preallocate slices, vectors, and maps when the final size is predictable                 | [Docs](docs/performance-by-default.md#perf_preallocate_collections)      | Go: `go/internal/analyzer/testdata/src/violations/violations.go`, Rust: `rust/fixtures/violations.rs` |
| `perf_avoid_reflection_dynamic` | Go, Rust  | Avoid reflection in Go and dynamic dispatch in Rust hot paths                            | [Docs](docs/performance-by-default.md#perf_avoid_reflection_dynamic)     | Go: `go/internal/analyzer/testdata/src/violations/violations.go`, Rust: `rust/fixtures/violations.rs` |
| `perf_bound_concurrency`        | Go, Rust  | Bound concurrency with worker pools or async limits to prevent oversubscription          | [Docs](docs/performance-by-default.md#perf_bound_concurrency)            | Go: `go/internal/analyzer/testdata/src/violations/violations.go`, Rust: `rust/fixtures/violations.rs` |
| `perf_borrow_instead_of_clone`  | Rust      | Prefer borrowing instead of cloning to avoid unnecessary allocations                     | [Docs](docs/performance-by-default.md#perf_borrow_instead_of_clone-rust) | Rust: `rust/fixtures/violations.rs`                                                                   |
| `perf_equal_fold_compare`       | Go        | Use `strings.EqualFold` instead of `strings.ToLower` / `strings.ToUpper` for comparisons | [Docs](docs/performance-by-default.md#perf_equal_fold_compare-go)        | Go: `go/internal/analyzer/testdata/src/violations/violations.go`                                      |
| `perf_vec_reserve_capacity`     | Rust      | Reserve capacity on vectors built inside deterministic loops                             | [Docs](docs/performance-by-default.md#perf_vec_reserve_capacity-rust)    | Rust: `rust/fixtures/violations.rs`                                                                   |
| `perf_syncpool_store_pointers`  | Go        | Store pointer types in `sync.Pool` to avoid interface allocation churn                   | [Docs](docs/performance-by-default.md#perf_syncpool_store_pointers-go)   | Go: `go/internal/analyzer/testdata/src/violations/violations.go`                                      |
| `perf_writer_prefer_bytes`      | Go        | Write byte slices directly instead of converting to strings                              | [Docs](docs/performance-by-default.md#perf_writer_prefer_bytes-go)       | Go: `go/internal/analyzer/testdata/src/violations/violations.go`                                      |
