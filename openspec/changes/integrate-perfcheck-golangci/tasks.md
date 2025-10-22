## 1. Discovery
- [ ] 1.1 Confirm GolangCI-Lint’s extension mechanisms and obtain maintainer approval for a built-in perfcheck linter.
- [ ] 1.2 Document compatibility targets (minimum GolangCI-Lint release, Go toolchain versions) and surface risks.

## 2. Implementation
- [ ] 2.1 Refactor perfcheck analyzers into an exported Go module with stable APIs for external consumers.
- [ ] 2.2 Add a GolangCI-Lint linter adapter that wires perfcheck analyzers into the aggregation pipeline with configuration support.
- [ ] 2.3 Update repository documentation to describe enabling perfcheck through `.golangci.yml` and remove redundant wrapper guidance.

## 3. Validation
- [ ] 3.1 Add automated tests (unit/integration) in GolangCI-Lint to cover perfcheck diagnostics and configuration parsing.
- [ ] 3.2 Run performance benchmarks to ensure the adapter meets perfcheck’s “performance-by-default” guarantees.
- [ ] 3.3 Pilot the integration on this repository and another external project to confirm real-world compatibility.

## 4. Release
- [ ] 4.1 Publish a perfcheck release tagged for GolangCI-Lint consumption and coordinate the corresponding GolangCI-Lint release notes.
- [ ] 4.2 Update support policy and changelog entries for both projects to highlight the new integration path.
