# Project Context

## Purpose
Build a cross-language "performance-by-default" linting toolkit that identifies inefficient patterns in Go and Rust code before they ship. Perfcheck codifies the research in `docs/` into reusable rules so engineering teams get fast feedback through static analysis rather than post-hoc profiling.

## Tech Stack
- Go 1.25.1 analyzer packaged as a `go vet`-compatible unitchecker (`go/cmd/perfcheck-go`).
- Rust 2021 linter library and CLI (`rust/`) that walks crates and emits diagnostics.
- Shared rule registry stored as TSV with JSON schema validation (`perfcheck-core/`).
- OpenSpec-driven workflow (`openspec/`) for tracking change proposals and specs.

## Project Conventions

### Code Style
- Follow `AGENTS.md` performance-first guidelines: small focused functions, explicit edge handling, and zero-allocation hot paths.
- Format code with toolchain defaults (`gofmt`, `rustfmt`); keep imports sorted and rule IDs stable (`perf_*`).
- Prefer numeric/token identifiers in analyzers, avoid string comparisons in tight loops, and reuse buffers aggressively.

### Architecture Patterns
- Central rule registry (`perfcheck-core/config/default_rules.tsv`) is the single source of truth; keep Go/Rust consumers in sync.
- Go analyzer embeds the TSV via `go/internal/ruleset` (run `go generate ./internal/ruleset` after TSV edits).
- Rust CLI includes the TSV at compile time with `include_str!` to avoid drift and emits `path:line:column [rule] message` diagnostics.
- Rules map to capability-specific detectors; add new languages by reusing the registry and providing language frontends.

### Testing Strategy
- Go: `cd go && GOCACHE=$(pwd)/.gocache go test ./...`; integration tests live alongside analyzers.
- Rust: `cd rust && cargo test`; ensure new lints have focused unit coverage.
- Update `docs/performance-by-default.md` and rule fixtures when adding or changing rules; keep cross-language behavior aligned.

### Git Workflow
- Favor short-lived feature branches named after the change ID; ensure commits reference the rule or change they touch.
- Simple branch model: feature branches off `main`, short-lived release/Hotfix branches when needed.
- Commits should follow Conventional Commits (`feat:`, `fix:`, `refactor:`, `test:`…) to keep changelog automation viable.
- Pull Requests must:
    - List configs/migrations touched and RPC-chain implications.
    - Attach test evidence (`go test`, `forge test`, benchmarks) or justify omissions.
    - Call out operational steps (deploy order, config promotions, indexer cache resets).
- Squash merges are preferred to maintain linear history per component while submodule updates track upstream revisions explicitly.


## Domain Context
- Focus on static analysis for performance antipatterns (e.g., string concatenation in loops, missing `reserve`/`with_capacity`).
- Go integration targets `go vet` workflows; Rust CLI is consumable directly or via embedding (`lint_source`, `lint_path`).
- Rule IDs double as numeric hashes for fast lookups—treat them as stable public identifiers.

## Important Constraints
- Avoid allocations, redundant state, and string comparisons in hot paths per `docs/performance-by-default.md`.
- Keep cross-language rule parity; whenever the TSV changes, regenerate Go bundles and ensure Rust continues to compile.
- Maintain deterministic tool outputs to support automated linting and spec validation.

## External Dependencies
- Go toolchain (`go build`, `go vet`) and Rust toolchain (`cargo`, `rustc`) for builds/tests.
- OpenSpec CLI for proposal/spec validation.
