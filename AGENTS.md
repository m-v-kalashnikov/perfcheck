<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

<!-- MANUAL ADDITIONS START -->
## Rules
- When creating new Rust modules, declare `mod foo;` / `pub mod foo;` in the parent, put the root in `foo.rs`, and place submodules under `foo/` with their own `mod` declarations (Rust 2018 layout). Never introduce `foo/mod.rs` unless fixing legacy code.
- Manage dependencies by declaring new crates in `[workspace.dependencies]` with `{ version = "...", default-features = false }`, and then enabling any required features per crate via `{ workspace = true, features = [...] }` without re-enabling defaults at the workspace level.
- When adding a new dependency, pick the latest stable release available (no `-alpha/-beta/-rc` unless explicitly required) and note any compatibility constraints if that is not possible.
- Check dependencies and remove any unused ones.
- Use testify for go testing
- Use nexttest for rust testing
<!-- MANUAL ADDITIONS END -->