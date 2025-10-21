# Tooling Integrations

This guide explains how to run perfcheck inside popular lint aggregators for
Go and Rust projects.

## GolangCI-Lint

1. Build the multichecker binary:
   ```bash
   cd go
   go build ./cmd/perfcheck-golangci
   ```
   Keep the resulting `perfcheck-golangci` binary somewhere on your `PATH` (for
   example `go/bin`).
2. Configure GolangCI-Lint to treat perfcheck as a custom external linter by
   adding the snippet below to `.golangci.yml`:
   ```yaml
   linters-settings:
     custom:
       perfcheck:
         cmd: perfcheck-golangci ./...
         description: Performance-by-default rules from perfcheck
         format: '{path}:{line}:{column}: {message}'

   linters:
     enable:
       - custom-perfcheck
   ```
   The multichecker emits diagnostics as `path:line:column [rule] message`, so
   GolangCI-Lint can parse them with the provided `format`.
3. Run `golangci-lint run` as usual; perfcheck violations now show up alongside
   built-in analyzers.

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
