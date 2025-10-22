## Why
- Developers lack an automated linting and audit workflow that covers Go and Rust code consistently.
- Missing tooling (golangci-lint, cargo fmt/clippy/deny/audit/udeps) allows style drift, latent warnings, and supply-chain risk to slip through CI.
- Aligning on a shared lint command lowers the barrier to running these checks locally and in automation.

## What Changes
- Introduce a documented Go workflow command that executes `golangci-lint run` against the repository with a checked-in configuration.
- Add a Rust maintenance workflow command that runs `cargo fmt --check`, `cargo clippy`, `cargo deny check`, `cargo audit`, and `cargo udeps` sequentially, exiting on the first failure.
- Update developer documentation and CI smoke guidance to cover the new lint/audit commands.

## Impact
- Additional tooling dependencies (`golangci-lint`, Rust cargo subcommands) must be installed in developer environments and CI images.
- Running the new commands will extend local validation time but catches style, lint, dependency, and security issues earlier.
- No breaking changes expected for existing smoke-test workflows; new commands complement the current test suite.
