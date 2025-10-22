## Why
Current Rust and Go maintenance workflows rely on a mix of tool defaults and ad-hoc local overrides, which makes lint and formatting output dependent on contributor toolchain installations. Clippy, rustfmt, Taplo, goimports, and GolangCI-Lint all support repository configuration files that we can pin to match the "performance by default" guidelines. Adding or tightening those configs ensures deterministic results and lets us codify lint groups that guard against perf regressions.

## What Changes
- Add `rust/clippy.toml` that elevates the performance-oriented lint groups we rely on and documents any intentional allow list.
- Add `rust/rustfmt.toml` to lock formatting behavior (module ordering, line width, edition) for the Rust workspace.
- Add `rust/taplo.toml` so TOML formatting stays stable across contributors and can back future automation.
- Encode goimports local prefix configuration so import grouping and local path detection stay consistent across developers (stored under `formatters.settings.goimports.local-prefixes` inside `go/.golangci.yml`).
- Update `go/.golangci.yml` to enforce the agreed performance-focused lint set and document any local-disable rationale.
- Document the configuration layout in the Rust tooling docs so contributors understand how to tweak rules safely.
- Update Go tooling documentation to describe how goimports and GolangCI-Lint consume the repository configs.

## Impact
- Rust linting/formatting commands produce consistent output regardless of local tool defaults.
- Go import formatting and lint output become deterministic across machines and CI.
- Contributors have a discoverable place to review or adjust lint decisions before they affect CI.
- No runtime impact; these are developer workflow improvements only.
