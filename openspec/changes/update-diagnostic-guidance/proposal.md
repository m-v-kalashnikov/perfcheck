## Why
Live usage surfaced that perfcheck diagnostics only echo a terse rule identifier and generic message. Engineers cannot see why the pattern is harmful or what to do next without hunting through docs, which slows adoption and reduces trust in the tool. We need actionable, built-in guidance so every reported issue immediately points to the fix.

## What Changes
- Extend the shared rule registry schema with per-rule explanation and fix-hint fields populated from our research source material.
- Update the Go analyzer diagnostic printer to stitch the rule id, short reason, and concrete remediation advice into every finding (CLI output and go/analysis diagnostics).
- Update the Rust linter diagnostic formatting and snapshot tests to emit the same explanation + fix hint payload alongside the rule id.
- Refresh docs/test fixtures to cover the richer messaging and ensure future rules provide guidance content by default.

## Impact
- Affected specs: perfcheck-core, go-analyzer, rust-linter, documentation (for messaging guidelines)
- Affected code: `perfcheck-core/config/default_rules.tsv`, `perfcheck-core/schema/rule_schema.json`, Go/Rust diagnostic formatting modules, analyzer/linter tests, docs explaining rule metadata
