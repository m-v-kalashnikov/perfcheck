## ADDED Requirements
### Requirement: Rust Tooling Configuration Baseline
The developer workflow SHALL version clippy, rustfmt, and Taplo configuration files inside the `rust/` workspace directory so that local and CI runs share identical lint and formatting behavior.

#### Scenario: Clippy honors repository configuration
- **WHEN** a contributor runs `cargo clippy` through the documented maintenance command
- **THEN** Clippy SHALL apply the rules specified in `rust/clippy.toml`, including any enforced deny lists.

#### Scenario: Rustfmt honors repository configuration
- **WHEN** a contributor runs `cargo fmt --check` via the maintenance workflow
- **THEN** rustfmt SHALL read `rust/rustfmt.toml` to determine formatting style, producing deterministic output across toolchain versions.

#### Scenario: Taplo honors repository configuration
- **WHEN** a contributor runs the documented `just rust-maintain` workflow
- **THEN** the Taplo step SHALL execute using `rust/taplo.toml` so configuration files and fixtures stay consistently formatted.

### Requirement: Go Tooling Configuration Baseline
The developer workflow SHALL version Go formatting and lint configuration inside the `go/` directory so that import formatting and lint behavior remain stable across contributors and CI.

#### Scenario: Goimports honors repository configuration
- **WHEN** a contributor formats Go files using the documented goimports command
- **THEN** goimports SHALL apply the repository-local import grouping (for example by invoking `goimports -local github.com/m-v-kalashnikov/perfcheck`) so local modules remain grouped consistently.

#### Scenario: GolangCI-Lint honors repository configuration
- **WHEN** a contributor runs the documented Go lint workflow
- **THEN** GolangCI-Lint SHALL source its configuration from the checked-in `go/.golangci.yml`, applying the enforced performance rule set and any documented exceptions.
