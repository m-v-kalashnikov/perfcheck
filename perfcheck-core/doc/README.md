# perfcheck-core documentation

This directory hosts reference material for the `perfcheck-core` package.
Populate it with content that helps integrators understand the rule registry
and configuration surface without digging through source files.

Suggested structure:

1. **Architecture overview** – Explain how the registry is loaded, how rule
   metadata is resolved from `config/default_rules.tsv`, and what guarantees
   consumers can rely on.
2. **Schema references** – Document the contract enforced by
   `schema/rule_schema.json`, including required fields, optional extensions,
   and examples of valid rule definitions.
3. **Extensibility guidance** – Capture how to introduce new rule families,
   version compatibility expectations, and any migration steps needed when
   schemas evolve.

Keep the docs synchronized with the implementation and update them whenever the
core APIs, configuration formats, or rule metadata change.
