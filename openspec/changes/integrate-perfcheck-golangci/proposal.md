## Why
Perfcheck can only run beside GolangCI-Lint today, forcing teams to maintain a separate vettool binary and wrapper scripts. Most Go shops standardise on GolangCI-Lint as their single linter entry point, so perfcheck’s current bridge limits adoption and adds friction to keep the analyzer in sync with development. Native GolangCI-Lint integration would let users enable perfcheck by configuration alone.

## What Changes
- Offer perfcheck as a first-class GolangCI-Lint linter that can be enabled in `.golangci.yml` without compiling an external vettool.
- Publish perfcheck’s analyzers as a reusable Go module that GolangCI-Lint can import while keeping rule IDs and performance guarantees intact.
- Update documentation and onboarding instructions to explain the new configuration-driven workflow and any compatibility constraints.
- Coordinate release and support processes so GolangCI-Lint consumers receive timely perfcheck updates.

## Impact
- Affected specs: `go-analyzer`
- Affected code: `go/internal/analyzer`, `go/cmd/perfcheck-golangci`, documentation under `docs/`, GolangCI-Lint downstream integration points
- External coordination: work with the GolangCI-Lint maintainers to upstream the new linter and define version compatibility expectations.
