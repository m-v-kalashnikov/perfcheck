## 1. Alignment & Guardrails
- [ ] 1.1 Confirm with GolangCI-Lint maintainers that upstreaming perfcheck as a built-in linter (vs plugin) is approved, capture required release window, and list the owning reviewers.
- [ ] 1.2 Finalize compatibility baselines per research: Go toolchain 1.24 for builds, GolangCI-Lint ≥ v2.6.0 (next minor), and perfcheck semantic versioning commitments; document risks if either side drifts.
- [ ] 1.3 Define support policy and escalation paths (perfcheck + GolangCI-Lint) covering bug reports, backports, and deprecation of the existing vettool wrapper.

## 2. Perfcheck Module Refactor
- [x] 2.1 Carve analyzers out of `internal/analyzer` into a public module (e.g. `go/pkg/perfchecklint`) that exports `Analyzers()` plus a `Build()` helper returning `goanalysis.Linter`.
- [x] 2.2 Expose stable rule metadata accessors (read-only registry, numeric rule IDs, severity defaults) so GolangCI-Lint can surface docs through `lintersdb`.
- [ ] 2.3 Retarget the module to Go 1.24, add CI to enforce the toolchain matrix, and publish the first tagged release (`v0.x.0`) with changelog + module documentation.

- [x] 3.1 Add the new module dependency to GolangCI-Lint (`go.mod/go.sum` + vendor) and wire it through Dependabot/security policies.
- [x] 3.2 Implement `pkg/golinters/perfcheck/perfcheck.go` that wraps the analyzers, plumbs configuration (rule allowlists, severity toggles), sets diagnostic categories, and registers metadata in `lintersdb`.
- [x] 3.3 Extend `pkg/config` and `.golangci.next.reference.yml` with perfcheck settings, provide fixtures under `pkg/golinters/perfcheck/testdata/`, and update `docs/linters/perfcheck.md` with usage instructions.
- [x] 3.4 Remove or deprecate the existing `cmd/perfcheck-golangci` wrapper inside this repo and replace documentation with the configuration-first workflow.

## 4. Validation & Quality Gates
- [ ] 4.1 Author unit and integration tests (go/analysis fixtures + GolangCI-Lint `pkg/golinters/tests`) covering rule diagnostics, configuration parsing, and compatibility with at least two sample projects.
- [ ] 4.2 Run and document performance benchmarks comparing the old vettool path vs the built-in linter, ensuring no regressions in runtime or allocations; gate merge on benchmark pass.
- [ ] 4.3 Execute pilot runs on this repository and one external codebase, capturing feedback/issues and feeding them back into the proposal before upstream submission.

## 5. Release & Communication
- [ ] 5.1 Coordinate simultaneous releases: tag the perfcheck module version consumed by GolangCI-Lint, update GolangCI-Lint release notes (`lintersdb` metadata + docs), and announce enablement instructions.
- [ ] 5.2 Update project changelogs, support policies, and onboarding guides to highlight the new integration pathway and removal timeline for the legacy vettool instructions.
