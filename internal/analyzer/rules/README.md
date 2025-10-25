# Analyzer rules layout

This directory is reserved for rule-specific Go packages that implement the
future modular analyzer architecture. Today the analyzer wires all rules
directly from `go/pkg/perfchecklint`, but the plan is to migrate each rule into
dedicated subpackages to keep logic cohesive and testable.

When introducing a new rule module:

1. Create a subdirectory (for example `perf_avoid_string_concat_loop`) with the
   analyzer implementation and focused unit tests.
2. Expose a constructor that returns a `go/analysis.Analyzer` so the registry
   can assemble the full rule set without bespoke wiring.
3. Ensure fixtures live under `go/pkg/perfchecklint/testdata/src` so they are
   reused by integration tests.

Once the migration is in place, this README should be updated with the finalized
layout and conventions.
