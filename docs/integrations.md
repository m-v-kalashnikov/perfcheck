# Tooling Integrations

This guide explains how to run perfcheck inside popular lint aggregators for
Go and Rust projects.

For this repository, the fastest path is to run the bundled commands:
- `just go-maintain` runs `golangci-lint fmt` (applying `gofmt`, `goimports`, `gci`, and `golines` with the repository settings), rebuilds the perfcheck vettool, runs `golangci-lint run` with the curated `go/.golangci.yml`, executes `go vet -vettool=./bin/perfcheck-go` across the module, checks `go.mod`, and runs `govulncheck ./...` (expect the first `govulncheck` run to download vulnerability data).
- `just rust-maintain` runs `cargo fmt --check`, validates all workspace TOML files with `taplo format --check --diff` using `rust/taplo.toml`, executes `cargo clippy --all-targets --all-features -- -D warnings`, `cargo deny check`, `cargo audit`, and `cargo +nightly udeps --all-targets` in sequence.
The Rust workflow requires `rustfmt`, `clippy`, `cargo-deny`, `cargo-audit`, and `cargo-udeps` along with a nightly toolchain; the deny/audit steps need the RustSec advisory database to be refreshed when network access is permitted.

## Repository Tooling Configuration

### Rust

- `rust/clippy.toml` documents the raw-string allowance used by our fixtures; lint levels are enforced through crate-level `#![warn(clippy::…)]` attributes in `src/lib.rs` and the executables. Any new `#![allow]` should explain why the rule is waived.
- `rust/rustfmt.toml` locks edition, width, and import grouping. Run `cargo fmt` to apply it across the workspace.
- `rust/taplo.toml` keeps TOML manifests deterministic. The `just rust-maintain` recipe calls `taplo format --check --diff` so format changes fail fast.

### Go

- `go/.golangci.yml` enables the performance-focused lint suite (`bodyclose`, `gocritic`, `prealloc`, `unconvert`) and configures `gofmt`, `goimports`, `gci`, and `golines` formatters. Manual runs should mirror CI with:
  - `goimports -local github.com/m-v-kalashnikov/perfcheck`
  - `gci write --custom-order --section standard --section default --section 'prefix(github.com/m-v-kalashnikov/perfcheck)' <files>`
  - `golines --max-len 120 --tab-len 8 <files>`
  The `gofmt` settings automatically rewrite `interface{}` to `any` so new code adopts modern type aliases by default. `just go-maintain` runs GolangCI-Lint with the repository cache directory so results remain deterministic.

## GolangCI-Lint

Perfcheck now publishes its analyzers and rule metadata through the
`github.com/m-v-kalashnikov/perfcheck/go/pkg/perfchecklint` module so
GolangCI-Lint can vendor the suite directly. Once the upstream release that
includes the built-in `perfcheck` linter ships (targeting GolangCI-Lint v2.6.0
or newer), repositories can enable it purely through configuration:

```yaml
linters:
  enable:
    - perfcheck

linters-settings:
  perfcheck:
    # Optional allowlist; omit to run all rules.
    rules:
      include:
        - perf_avoid_string_concat_loop
    # Optional severity overrides.
    severity:
      perf_prefer_stack_alloc: warning
```

Perfcheck follows the GolangCI-Lint go/analysis adapter conventions:

- Rule IDs map 1:1 with `perfchecklint.Rules()` metadata so `nolint:perf_*`
  continues to work.
- GolangCI-Lint enforces the compatibility matrix documented in the proposal
  (Go toolchain ≥1.24, GolangCI-Lint ≥v2.6.0). When the versions drift, the
  perfcheck linter fails fast with an actionable error.
- `linters-settings.perfcheck.rules.include` can be used to restrict emission to
  a subset of rule identifiers, and future releases will add severity toggles in
  the same block.

Until the upstream release is available, keep running the analyzers directly via
`go vet -vettool=$(pwd)/perfcheck-go ./...` (see the Go analyzer section above).

## Clippy

1. Install the perfcheck binaries so `cargo` can find both the standalone CLI
   and the Clippy bridge:
   ```bash
   cargo install --path rust --bin perfcheck --bin cargo-perfcheck-clippy
   ```
2. Invoke the combined workflow with:
   ```bash
   cargo perfcheck-clippy
   ```
   The subcommand first runs `cargo clippy` with all original arguments and,
   when Clippy succeeds, launches the perfcheck CLI on the same crate.

The runner accepts an optional `--perfcheck-target=<path>` (or the split form
`--perfcheck-target <path>`) to point at a specific crate directory. If absent,
it falls back to the directory containing `--manifest-path` and finally to the
current working directory. You can add an alias or CI job that calls
`cargo perfcheck-clippy --all-targets -- -D warnings` to enforce both lint sets
in lockstep.
