## Why
PerfCheck currently lacks coverage for several high-impact inefficiency patterns that show up repeatedly in the "Proposed PerfCheck Rule Candidates" research doc. These gaps include misuse of linked lists, mutexes guarding primitives, defer statements in loops, needless rune conversions, unbuffered I/O, oversized enum variants, redundant Arc usage, wasteful collect calls, and gratuitous heap allocations for tiny objects. Without codifying these findings into specs, the analyzers cannot warn users about them.

## What Changes
- Expand the shared rule registry with nine new rule IDs derived from the candidate list so every analyzer can surface consistent diagnostics.
- Update the Go analyzer specification so it must detect linked-list usage, defer-in-loop patterns, rune slice conversions, unbuffered I/O, mutexes guarding primitives, and needless heap allocations of small structs/values.
- Update the Rust linter specification so it must detect linked-list usage, large enum variants, redundant Arc<T>, mutexes guarding primitives, needless collect calls, and unnecessary heap indirection for small Copy types.
- Document that both analyzers share the new metadata via the core registry to keep rule guidance in sync.

## Impact
- Affected specs: perfcheck-core, go-analyzer, rust-linter
- Affected code: TSV registry in `perfcheck-core`, Go analyzer detectors (`go/internal/...`), Rust linter rules (`rust/src/rules/...`), documentation describing rule catalogue
