# Performance-by-Default Methodology

This document summarizes coding principles that make software efficient **by design**.

## Key Principles
1. **Minimize Allocations**
   - Go: preallocate slices/maps, reuse buffers, prefer sync.Pool.
   - Rust: use borrowing instead of cloning, with_capacity, stack data.

2. **Prefer Stack & Preallocation**
   - Go: escape analysis — keep values on stack when possible.
   - Rust: stack allocation is default; heap only when necessary.

3. **Avoid Work Inside Hot Loops**
   - Hoist invariant code outside loops.
   - Avoid fmt.Sprintf or String += in loops (use builders).

4. **Batch Expensive Operations**
   - Combine syscalls, I/O, and cgo/FFI calls into fewer, larger calls.

5. **Data-Structure Awareness**
   - Favor contiguous memory (slices, Vec).
   - Use maps only when needed; avoid string keys in Go unless hashed.

6. **Concurrency with Limits**
   - Go: use worker pools, buffered channels, context cancellation.
   - Rust: prefer Rayon or async executors; avoid excessive threads.

7. **Static Analysis Enforcement**
   - Go: integrate rules similar to SA6000–SA6006, gocritic.hugeParam, etc.
   - Rust: enforce Clippy rules (needless_clone, needlessly_pass_by_value, string_add).

8. **Profile & Measure**
   - Use pprof (Go) or Criterion (Rust) before and after changes.

> **Rule IDs:** Canonical identifiers are maintained in `perfcheck-core/config/default_rules.tsv` (prefixed `perf_*`). Keep documentation snippets aligned with those IDs.

See perfcheck-core/config/default_rules.tsv for the canonical rule registry.

## Analyzer Examples

### `perf_avoid_string_concat_loop`
#### Go
```go
func join(items []string) string {
    out := ""
    for _, item := range items {
        out += item // perf_avoid_string_concat_loop: use strings.Builder instead
    }
    return out
}
```

#### Rust
```rust
fn join(items: &[String]) -> String {
    let mut out = String::new();
    for item in items {
        out += item; // perf_avoid_string_concat_loop: prefer String::with_capacity + push_str
    }
    out
}
```

### `perf_regex_compile_once` (Go)
```go
func matchAll(inputs []string, expr string) int {
    count := 0
    for _, text := range inputs {
        if regexp.MustCompile(expr).MatchString(text) { // perf_regex_compile_once: hoist MustCompile outside
            count++
        }
    }
    return count
}
```

### `perf_preallocate_collections`
#### Go
```go
func build(items []string) []string {
    out := make([]string, 0, len(items)) // reserve final length up front
    for _, item := range items {
        out = append(out, item)
    }
    return out
}
```

#### Rust
```rust
fn collect(count: usize) -> Vec<i32> {
    let mut data = Vec::with_capacity(count); // reserve before the loop
    for idx in 0..count {
        data.push(idx as i32);
    }
    data
}
```

### `perf_avoid_reflection_dynamic`
#### Go
```go
func types(values []any) []reflect.Kind {
    kinds := make([]reflect.Kind, 0, len(values))
    for _, v := range values {
        kinds = append(kinds, reflect.TypeOf(v).Kind()) // perf_avoid_reflection_dynamic: cache reflection outside
    }
    return kinds
}
```

#### Rust
```rust
trait Handler {
    fn handle(&self, value: i32);
}

fn process(items: &[i32], handler: &dyn Handler) {
    for value in items {
        handler.handle(*value); // perf_avoid_reflection_dynamic: avoid dyn dispatch in the loop
    }
}
```

### `perf_bound_concurrency`
#### Go
```go
func process(tasks []func()) {
    for _, task := range tasks {
        go task() // perf_bound_concurrency: guard with a worker pool or semaphore
    }
}
```

#### Rust
```rust
use std::thread;

fn process(items: &[i32]) {
    for item in items {
        thread::spawn(move || println!("{}", item)); // perf_bound_concurrency: spawn inside a loop without bounds
    }
}
```

### `perf_borrow_instead_of_clone` (Rust)
```rust
fn sum_lengths(values: &[String]) -> usize {
    let mut total = 0;
    for value in values {
        total += value.clone().len(); // perf_borrow_instead_of_clone: borrow `value` instead of cloning
    }
    total
}
```

### `perf_equal_fold_compare` (Go)
```go
func equalInsensitive(a, b string) bool {
    return strings.ToLower(a) == strings.ToLower(b) // perf_equal_fold_compare: use strings.EqualFold
}
```

### `perf_vec_reserve_capacity` (Rust)
```rust
fn collect(values: &[i32]) -> Vec<i32> {
    let mut out = Vec::new();
    for value in values {
        out.push(*value); // perf_vec_reserve_capacity: reserve capacity before pushing
    }
    out
}
```

### `perf_syncpool_store_pointers` (Go)
```go
type item struct{ payload [1024]byte }

func store(pool *sync.Pool, value item) {
    pool.Put(value) // perf_syncpool_store_pointers: store *item instead to avoid copying
}
```

### `perf_writer_prefer_bytes` (Go)
```go
func writeBytes(w io.Writer, buf []byte) {
    io.WriteString(w, string(buf)) // perf_writer_prefer_bytes: write the []byte directly
}
```

## Validation Workflow
- Run `just go-maintain` to apply `golangci-lint fmt` (wrapping `gofmt`, `goimports`, `gci`, and `golines`), compile the GolangCI-Lint bridge, enforce the analyzer suite (including `testifylint`, `wastedassign`, and `whitespace`), verify modules, and ensure `govulncheck ./...` reports no vulnerabilities (first run may download advisory data).
- Run `just rust-maintain` to verify formatting, clippy diagnostics, supply-chain checks, and unused dependency drift (requires installed `cargo-deny`, `cargo-audit`, and a nightly toolchain for `cargo udeps`; keep the RustSec database synced when network access is available).
- Run `just pre-commit` to execute the lint commands, `go test ./...`, `cargo nextest run`, and `openspec validate --strict` in one pass before submitting changes.
