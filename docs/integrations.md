# Tooling Integrations

This guide explains how to run perfcheck inside popular lint aggregators for
Go and Rust projects.

For this repository, the fastest path is to run the bundled commands:
- `just lint-go` builds the GolangCI-Lint bridge, runs `golangci-lint run` with the checked-in configuration, and then executes the perfcheck multichecker across the Go sources.
- `just maintain-rust` runs `cargo fmt --check`, `cargo clippy --all-targets --all-features -- -D warnings`, `cargo deny check`, `cargo audit`, and `cargo +nightly udeps --all-targets` in sequence.
The Rust workflow requires `rustfmt`, `clippy`, `cargo-deny`, `cargo-audit`, and `cargo-udeps` along with a nightly toolchain; the deny/audit steps need the RustSec advisory database to be refreshed when network access is permitted.

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
   The `just lint-go` recipe automates both steps when working inside this repository.

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
