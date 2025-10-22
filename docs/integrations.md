# Tooling Integrations

This guide explains how to run perfcheck inside popular lint aggregators for
Go and Rust projects.

For this repository, the fastest path is to run the bundled commands:
- `just go-maintain` runs `golangci-lint fmt` (applying `gofmt`, `goimports`, `gci`, and `golines` with the repository settings), builds the GolangCI-Lint bridge, runs `golangci-lint run` with the curated `go/.golangci.yml`, executes the perfcheck multichecker, checks `go.mod`, and runs `govulncheck ./...` (expect the first `govulncheck` run to download vulnerability data).
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

1. Build the multichecker binary:
   ```bash
   cd go
   go build ./cmd/perfcheck-golangci
   ```
   Keep the resulting `perfcheck-golangci` binary somewhere on your `PATH` (for
   example `go/bin`).
2. Run `golangci-lint run` as usual using `go/.golangci.yml`; this covers the standard static analyzers.
3. Execute the perfcheck bridge separately to surface performance violations:
   ```bash
   ./bin/perfcheck-golangci ./...
   ```
  The `just go-maintain` recipe automates these steps and adds module verification and `govulncheck` when working inside this repository.

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
