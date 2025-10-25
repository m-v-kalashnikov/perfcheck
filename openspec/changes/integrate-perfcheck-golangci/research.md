# integrate-perfcheck-golangci Research

## Objectives
- Document how GolangCI-Lint loads third-party analyzers so we can register Perfcheck as a first-class linter without wrappers.
- Capture packaging and compatibility requirements (Go toolchain, GolangCI-Lint releases, module APIs).
- Identify testing, release, and support expectations from both the Perfcheck and GolangCI-Lint sides.

## Current State Snapshot
- Perfcheck exposes analyzers only through `internal/analyzer` and ships a `cmd/perfcheck-golangci` multichecker vettool. External consumers cannot import `internal/...` packages, so GolangCI-Lint cannot embed the analyzers today.
- Rule metadata lives in `internal/ruleset` and is bundled into the vettool binary; there is no stable exported API or semantic versioning for analyzers.
- Repositories run Perfcheck by building the vettool and wiring GolangCI-Lint’s `run` into `--out-format` parsers. This duplicates configuration and makes upgrades hard.

## GolangCI-Lint Integration Surface
GolangCI-Lint v2.5.0 (latest release as of https://github.com/golangci/golangci-lint/releases/tag/v2.5.0) offers two extension paths (per https://golangci-lint.run/docs/contributing/new-linters/):
1. **Upstream (preferred)** – contribute a new linter into the `golangci/golangci-lint` repo. Requirements:
   - Implement `golangci-lint/pkg/golinters/base.GoLinter` (usually via helpers in `pkg/goanalysis`).
   - Drop implementation under `pkg/golinters/perfcheck/perfcheck.go` and add smoke fixtures under `pkg/golinters/perfcheck/testdata/perfcheck/perfcheck.go`.
   - Register the linter in `pkg/lint/lintersdb/builder_linter.go` (`LinterBuilder.Build`) with `WithSince("vX.Y.Z")` pointing at the next GolangCI-Lint minor release and continue populating metadata in `pkg/lintersdb/db.go` so it appears in `golangci-lint linters`.
   - Add configuration struct under `pkg/config` (with `mapstructure` tags) and surface defaults in `.golangci.next.reference.yml` (plus `.golangci.yml` if we want non-default behavior in the repo).
   - Provide unit/integration tests (`pkg/golinters/tests/...`) and update documentation under `docs/linters/`.
   - Follow the documented validation command: `go run ./cmd/golangci-lint/ run --no-config --default=none --enable=perfcheck ./pkg/golinters/perfcheck/testdata/perfcheck.go`.
2. **Plugin (not viable)** – build a Go plugin and load at runtime. This is Linux-only, brittle across Go versions, and explicitly discouraged by GolangCI-Lint maintainers for production linters.

Given the proposal’s "first-class" requirement, we must follow path 1 and upstream Perfcheck directly.

### go/analysis Adapter Expectations
- GolangCI-Lint wraps go/analysis analyzers via `goanalysis.NewLinter(name, description, analyzers...)`.
- Rules emit diagnostics through `analysis.Diagnostic`. GolangCI-Lint maps `Diagnostic.Category` (or custom facts) into `Issue.RuleID`. We must ensure every analyzer sets `Diagnostic.Category` (e.g., `perf_preallocate_collections`).
- Performance classification: in `lintersdb.Linter`, set `IsSlow`/`IsExperimental` flags to help users.
- Configuration: add `Perfcheck` struct under `pkg/config/linters_settings.go`, expose knobs such as `rules`, `severity`, or `experimental`. Documentation must describe defaults.

## Packaging Strategy for Perfcheck
1. **Create Public Package** – move analyzers from `internal/analyzer` into a public module path, e.g. `github.com/m-v-kalashnikov/perfcheck/go/pkg/perfchecklint`.
   - Export: `func Analyzers() []*analysis.Analyzer`, `func Build(name string, opts Options) *goanalysis.Linter`, and structs for configuration (rule allowlists, severity mapping, etc.).
   - Keep `internal` for rule metadata loading; expose read-only views instead of allowing callers to mutate state.
2. **Versioning** – tag releases `v0.x.0` in the Go module to give GolangCI-Lint a stable semantic version. Document compatibility matrix (see below).
3. **Go Toolchain** – module currently declares `go 1.25`. GolangCI-Lint v2.5.0 binaries are built with Go 1.24.0 (per release artifacts), so we need to either:
   - Retarget Perfcheck’s module to Go 1.24 to match the upstream build environment, or
   - Provide build tags / shims so GolangCI-Lint can vendor Perfcheck without downgrading its compiler.
   Action: align on **Go 1.24** support until GolangCI-Lint publishes a newer toolchain baseline.
4. **Rule Metadata Access** – provide helper API so GolangCI-Lint can surface rule docs (name, summary, remediation) in `golangci-lint linters --json` output. Example: `func Registry() map[string]RuleDoc` returning copies of metadata.

## GolangCI-Lint Code Touch Points
| File/Area | Purpose | Required Change |
|-----------|---------|------------------|
| `go.mod` / `go.sum` | Add dependency on `github.com/m-v-kalashnikov/perfcheck/go/pkg/perfchecklint`. | Pin to released tag; enable Dependabot rules. |
| `pkg/golinters/perfcheck/perfcheck.go` (new) | Adapter that wraps Perfcheck analyzers using `goanalysis.NewLinter`. | Wire configuration (rule filters, severity) and expose WithSince metadata. |
| `pkg/golinters/perfcheck/testdata/` | Functional fixtures required by GolangCI-Lint contribution guide. | Provide representative sample code covering at least one diagnostic per analyzer. |
| `pkg/lint/lintersdb/builder_linter.go` | Register linter so it shows up in `golangci-lint linters`. | Call `WithSince("v2.6.0")` (or appropriate next minor) and hook `perfcheck` metadata. |
| `pkg/lintersdb/db.go` | Register metadata (name `perfcheck`, description, presets). | Mark `HasSettings: true`, `IsSlow: true/false` (estimate), `OriginalURL`. |
| `.golangci.next.reference.yml` & `.golangci.yml` | Showcase configuration surface per upstream doc. | Add example `linters-settings.perfcheck` block with sane defaults. |
| `pkg/config/config.go` / `linters_settings.go` | Define `Perfcheck` settings struct; add YAML wiring so `.golangci.yml` can control Perfcheck. | Include validation (e.g., unknown rule IDs) and `mapstructure` tags. |
| `docs/linters/perfcheck.md` + `README` | Document enabling instructions and compatibility table. | Provide sample `.golangci.yml`. |
| `pkg/golinters/perfcheck/perfcheck_test.go` | Add smoke tests verifying diagnostics appear and config parsing works. | Use `testdata/src/perfcheck/...`. |
| `pkg/app/runners.go` (if necessary) | Ensure new linter is included in default linter set. | Typically no changes beyond registration. |

## Configuration Contract
Suggested `.golangci.yml` snippet:
```yaml
linters:
  enable:
    - perfcheck
linters-settings:
  perfcheck:
    rules:
      include: [perf_preallocate_collections, perf_defer_in_loop]
      exclude: [perf_equal_fold]
    fail-on-unsupported-version: true
```
Implementation details:
- `rules.include/exclude` allow teams to scope detectors without recompiling.
- Add `min_go_version` check so Perfcheck can skip analyzers needing generics metadata.
- Provide `severity` mapping (e.g., `warn` vs `error`).
- Ensure `.golangci.next.reference.yml` includes the block so users discover knobs via upstream docs.

## Compatibility Matrix (Initial Proposal)
| Perfcheck Module | GolangCI-Lint Release | Go Toolchain | Notes |
|------------------|-----------------------|--------------|-------|
| `v0.3.x`         | `v2.5.0+`             | Go 1.24.x    | First release with exported `perfchecklint` package; aligns with latest GolangCI-Lint.
| `v0.4.x`         | `v2.6.0+` *(planned)* | Go 1.25.x    | Adds analyzers that rely on new go/types APIs; keep guards when running on older Go.
| `v0.5.x`         | `v2.8.0+` *(planned)* | Go 1.26.x    | Requires future GolangCI-Lint toolchain bump; document migration steps.

Action items:
- Publish matrix in Perfcheck docs and GolangCI-Lint release notes.
- Implement runtime guard in adapter: if linked Perfcheck version < required, emit fatal error with remediation link.

## Testing & Validation Expectations
1. **Unit tests inside Perfcheck** – keep existing analyzer tests and add coverage for exported API (rule registry, options parsing).
2. **GolangCI-Lint integration tests** – follow existing pattern (`pkg/golinters/tests`). Add cases:
   - `linters-settings.perfcheck.rules.include` filters diagnostics.
   - Version mismatch surfaces actionable error.
   - Perfcheck respects `run.exclude-dirs` and per-file `nolint` directives.
   - Run the doc-mandated validation command: `go run ./cmd/golangci-lint/ run --no-config --default=none --enable=perfcheck ./pkg/golinters/perfcheck/testdata/perfcheck.go`.
3. **Performance validation** – run `golangci-lint run --enable perfcheck ./...` on:
   - This repo (baseline) – ensure runtime within 10% of current vettool.
   - Another open-source Go project (~1–2k files) – confirm no excessive memory growth (<512 MiB) and runtime <2× baseline.
4. **Pilot** – vendor pre-release GolangCI-Lint binary with Perfcheck enabled, run in CI for Perfcheck repo and at least one partner project.

## Release & Support Notes
- Perfcheck releases must be tagged before GolangCI-Lint can update `go.mod`. Establish process: cut Perfcheck tag ➝ update GolangCI-Lint dependency ➝ add release notes.
- Provide CHANGELOG section "GolangCI-Lint consumers" summarizing new/removed rules and config migrations.
- Coordinate with GolangCI-Lint maintainers via GitHub issue/PR; they expect:
  - Linter owners triage bugs filed in GolangCI-Lint repo that are analyzer-specific.
  - Compatibility policy (e.g., support latest two Go releases).
- Document fallback strategy for users pinned to older GolangCI-Lint (keep vettool for at least one release cycle).

## Risks & Mitigations
| Risk | Impact | Mitigation |
|------|--------|-----------|
| Toolchain mismatch (Go 1.25 modules vs GolangCI-Lint 1.24) | Build failure when vendoring Perfcheck. | Re-target Perfcheck module to Go 1.24 or add build tags per analyzer. |
| Analyzer memory usage increases GolangCI-Lint runtime > acceptable. | Users disable linter due to slowness. | Profile analyzers inside GolangCI-Lint runner; cache rule metadata; avoid allocations in hot loops. |
| Rule ID drift between vettool and GolangCI-Lint. | Confusing diagnostics / `nolint` directives break. | Centralize rule metadata; enforce tests that GolangCI-Lint adapter uses same IDs. |
| Upstream review delays. | Integration misses release window. | Open design issue with GolangCI-Lint maintainers early; share perf data; keep PR small. |
| Config UX confusion (two rule filters). | Users misconfigure. | Provide docs + validation errors for unknown rule IDs.

## References
- GolangCI-Lint contributing docs ("Adding New Linters", https://golangci-lint.run/docs/contributing/new-linters/)
- GolangCI-Lint goanalysis helper (pkg/goanalysis in github.com/golangci/golangci-lint)
- Perfcheck internal analyzer package (`internal/analyzer` in this repo)
- Proposal/tasks under `openspec/changes/integrate-perfcheck-golangci/`
