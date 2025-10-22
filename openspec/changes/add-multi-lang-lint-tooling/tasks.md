## 1. Go Lint Workflow
- [x] 1.1 Add repository-level configuration for `golangci-lint` covering existing performance analyzers and desired rules.
- [x] 1.2 Expose a documented command (script or make target) that runs `golangci-lint run` from the repository root and fails on lint findings.

## 2. Rust Maintenance Workflow
- [x] 2.1 Script a rust workflow command that sequentially runs `cargo fmt --check`, `cargo clippy --all-targets --all-features -D warnings`, `cargo deny check`, `cargo audit`, and `cargo udeps --all-targets` with fast-fail behavior.
- [x] 2.2 Ensure required cargo subcommands are added to developer setup instructions or tooling bootstrap scripts.

## 3. Documentation & Automation
- [x] 3.1 Update developer workflow docs to describe when to run the Go and Rust lint/audit commands.
- [x] 3.2 Integrate the new commands into CI or the documented smoke-test suite so regressions surface automatically.
- [x] 3.3 Validate by running both commands locally and capturing their passing output in the change log or PR notes. *(Executed `just lint-go` and `just maintain-rust`; both succeeded.)*
