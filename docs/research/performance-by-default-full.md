---
title: "Performance-by-Default Methodology — Go and Rust"
author: "Michael Kalashnikov"
date: "2025-10-21"
version: "1.0"
description: "Comprehensive research defining Performance-by-Default methodology for Go and Rust, including rules, examples, references, and static analysis mappings."
---

# Performance-by-Default Methodology — Go & Rust

This document defines a **Performance-by-Default** methodology — a cross-language framework for writing software that is efficient by design.  
It integrates **Go** and **Rust** performance patterns, anti-patterns, and static analysis mappings suitable for automated linting in `perfcheck`.

> **Scope:** Practical, code-first guidance with inline examples and links. Each rule block includes metadata (id, langs, category, severity) and static analysis mappings (Go staticcheck/gocritic, Rust Clippy) where applicable.

---

## 0) Core Principles

- **Minimize allocations**: prefer stack, preallocate, reuse.  
- **Prefer contiguous data**: slices/Vec over linked lists.  
- **Move invariants out of hot loops**.  
- **Batch expensive ops**: I/O, syscalls, FFI/cgo.  
- **Bound concurrency**: pools/queues; avoid explosion.  
- **Avoid runtime reflection / dynamic dispatch in hot paths**.  
- **Measure**: benchmark & profile before/after changes.

---

## 1) Go — Performance-by-Default

### 1.1 Minimize Heap Allocations
```yaml
id: go_minimize_heap_allocations
langs: [go]
category: allocation
severity: warning
```
**Why:** Every heap allocation adds GC pressure.  
**Do:** Preallocate and reuse.

**Bad**
```golang
func build() []int {
    var out []int
    for i := 0; i < 10_000; i++ {
        out = append(out, i) // multiple resizes
    }
    return out
}
```

**Good**
```golang
func build() []int {
    out := make([]int, 0, 10_000) // pre-capacity
    for i := 0; i < 10_000; i++ {
        out = append(out, i)
    }
    return out
}
```

- 🔍 Go: consider gocritic `hugeParam` for pass-by-value size hints: https://go-critic.github.io/overview.html#hugeparam-ref  
- 📖 Effective Go (allocation efficiency): https://go.dev/doc/effective_go#allocation_efficiency

---

### 1.2 Avoid Reflection in Hot Paths
```yaml
id: go_avoid_reflection_hotpaths
langs: [go]
category: runtime
severity: high
```
**Why:** `reflect` incurs dynamic checks and can allocate.

**Bad**
```golang
import "reflect"

func kindOf(v any) string {
    return reflect.ValueOf(v).Kind().String()
}
```

**Good**
```golang
// Prefer static types and interfaces with concrete impls for hot paths.
type Kinded interface{ Kind() string }

func kindOf(v Kinded) string { return v.Kind() }
```

- 🔍 Go: reflection is not directly linted as “perf”, but prefer compile-time typing.  
- 📖 Laws of Reflection: https://go.dev/blog/laws-of-reflection

---

### 1.3 String Concatenation in Loops
```yaml
id: go_avoid_string_concat_in_loops
langs: [go]
category: memory
severity: warning
```
**Why:** Each `+=` copies & allocates a new string.

**Bad**
```golang
func join(xs []string) string {
    s := ""
    for _, x := range xs {
        s += x // alloc each iteration
    }
    return s
}
```

**Good**
```golang
import "strings"

func join(xs []string) string {
    var b strings.Builder
    for _, x := range xs {
        b.WriteString(x)
    }
    return b.String()
}
```

- 📖 strings.Builder: https://go.dev/blog/strings

---

### 1.4 Case-Insensitive Compare
```yaml
id: go_case_insensitive_compare
langs: [go]
category: string
severity: low
```
**Bad**
```golang
if strings.ToLower(a) == strings.ToLower(b) { ... }
```

**Good**
```golang
if strings.EqualFold(a, b) { ... }
```
- 🔍 Go: staticcheck **SA6005**: https://staticcheck.io/docs/checks#SA6005

---

### 1.5 Byte Slice → String for Map Key
```yaml
id: go_inline_byte_to_string_mapkey
langs: [go]
category: allocation
severity: medium
```
**Why:** Assigning `tmp := string(b)` allocs; `m[string(b)]` in index context can avoid it.

**Bad**
```golang
k := string(b)
v := m[k] // missed optimization
```

**Good**
```golang
v := m[string(b)] // zero-alloc fast path in this context
```
- 🔍 Go: staticcheck **SA6001**: https://staticcheck.io/docs/checks#SA6001

---

### 1.6 sync.Pool: Store Pointers
```yaml
id: go_syncpool_store_pointers
langs: [go]
category: allocation
severity: medium
```
**Bad**
```golang
var bufPool = sync.Pool{New: func() any { return make([]byte, 0, 4096) }}
// storing []byte as interface can cause extra allocs in some flows
```

**Good**
```golang
var bufPool = sync.Pool{New: func() any { b := make([]byte, 0, 4096); return &b }}
// store &[]byte (or *bytes.Buffer)
```
- 🔍 Go: staticcheck **SA6002**: https://staticcheck.io/docs/checks#SA6002

---

### 1.7 Efficient I/O Writes
```yaml
id: go_avoid_write_string_for_bytes
langs: [go]
category: io
severity: low
```
**Bad**
```golang
io.WriteString(w, string(buf)) // alloc string
```

**Good**
```golang
w.Write(buf) // no allocation
```
- 🔍 Go: staticcheck **SA6006**: https://staticcheck.io/docs/checks#SA6006

---

### 1.8 Avoid Regex Compilation in Loops
```yaml
id: go_regex_compile_once
langs: [go]
category: cpu
severity: medium
```
**Bad**
```golang
for _, s := range data {
    re := regexp.MustCompile(`\w+@\w+\.\w+`)
    _ = re.MatchString(s)
}
```

**Good**
```golang
var re = regexp.MustCompile(`\w+@\w+\.\w+`)
for _, s := range data {
    _ = re.MatchString(s)
}
```
- 🔍 Go: staticcheck **SA6000**: https://staticcheck.io/docs/checks#SA6000

---

### 1.9 Bound Concurrency (Avoid Goroutine Explosion)
```yaml
id: go_bounded_concurrency
langs: [go]
category: concurrency
severity: critical
```
**Bad**
```golang
for _, job := range jobs {
    go handle(job) // unbounded
}
```

**Good (worker pool)**
```golang
sem := make(chan struct{}, 128)
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    sem <- struct{}{}
    go func(j Job) {
        defer wg.Done()
        handle(j)
        <-sem
    }(job)
}
wg.Wait()
```

- 📖 Pipelines: https://go.dev/blog/pipelines

---

### 1.10 Move Invariants Out of Loops
```yaml
id: go_hoist_invariants
langs: [go]
category: cpu
severity: info
```
**Bad**
```golang
sum := 0
for i := 0; i < len(xs); i++ { // len repeatedly
    sum += xs[i]
}
```

**Good**
```golang
n := len(xs)
sum := 0
for i := 0; i < n; i++ {
    sum += xs[i]
}
```

---

### 1.11 Avoid Linked Lists for Traversal
```yaml
id: go_prefer_slices_over_lists
langs: [go]
category: cache
severity: info
```
**Why:** `container/list` allocates per node; poor locality. Prefer slices.

- 📖 Slice tricks: https://dave.cheney.net/2018/07/12/slices-from-the-ground-up

---

### 1.12 Escape Analysis Awareness
```yaml
id: go_escape_analysis_awareness
langs: [go]
category: allocation
severity: info
```
Use `go build -gcflags="-m"` to inspect escapes; adjust code to keep values on stack when possible.  
- 📖 Escape analysis keynote: https://go.dev/blog/ismmkeynote

---

### 1.13 Batch Syscalls and cgo
```yaml
id: go_batch_syscalls_cgo
langs: [go]
category: io
severity: medium
```
Buffer I/O and batch cgo work to minimize boundary crossings.  
- 📖 cgo performance: https://go.dev/doc/cgo#perf

---

### 1.14 PGO & Up-to-date Runtime
```yaml
id: go_enable_pgo_and_update
langs: [go]
category: compiler
severity: info
```
Use PGO (Go 1.21+) and upgrade Go for ongoing compiler/runtime optimizations.  
- 📖 Go 1.21: https://go.dev/doc/go1.21

---

## 2) Rust — Performance-by-Default

### 2.1 Borrow Instead of Clone
```yaml
id: rs_borrow_not_clone
langs: [rust]
category: allocation
severity: high
```
**Bad**
```rust
fn count_names(names: Vec<String>) -> usize {
    let mut total = 0;
    for n in names.clone() { // unnecessary clone
        total += n.len();
    }
    total
}
```

**Good**
```rust
fn count_names(names: &[String]) -> usize {
    let mut total = 0;
    for n in names { // borrow
        total += n.len();
    }
    total
}
```

- 🦀 Clippy: `redundant_clone` https://rust-lang.github.io/rust-clippy/master/#redundant_clone  
- 📖 Ownership: https://doc.rust-lang.org/book/ch04-01-what-is-ownership.html

---

### 2.2 Prefer &str over String params
```yaml
id: rs_prefer_str_param
langs: [rust]
category: allocation
severity: medium
```
**Bad**
```rust
fn greet(name: String) -> String {
    format!("Hello, {name}!")
}
```

**Good**
```rust
fn greet(name: &str) -> String {
    format!("Hello, {name}!")
}
```
- 🦀 Clippy: `ptr_arg` https://rust-lang.github.io/rust-clippy/master/#ptr_arg

---

### 2.3 Reserve Capacity for Vec/String
```yaml
id: rs_reserve_capacity
langs: [rust]
category: allocation
severity: warning
```
**Bad**
```rust
let mut v = Vec::new();
for i in 0..10_000 {
    v.push(i);
}
```

**Good**
```rust
let mut v = Vec::with_capacity(10_000);
for i in 0..10_000 {
    v.push(i);
}
```

- 📖 Vec::with_capacity: https://doc.rust-lang.org/std/vec/struct.Vec.html#method.with_capacity

---

### 2.4 Avoid LinkedList for Traversal
```yaml
id: rs_avoid_linkedlist_for_traversal
langs: [rust]
category: cache
severity: info
```
Prefer `Vec` for better cache locality.  
- 📖 Rust Performance Book (Data structures): https://nnethercote.github.io/perf-book/  

---

### 2.5 Use Iterators (Zero-Cost)
```yaml
id: rs_prefer_iterators
langs: [rust]
category: cpu
severity: info
```
**Bad**
```rust
let mut sum = 0;
for i in 0..v.len() { sum += v[i]; } // potential bounds checks
```

**Good**
```rust
let sum: i32 = v.iter().copied().sum(); // fused, no bounds checks
```
- 📖 Iterators: https://nnethercote.github.io/perf-book/iterators.html

---

### 2.6 Static Dispatch in Hot Paths
```yaml
id: rs_prefer_static_dispatch
langs: [rust]
category: cpu
severity: medium
```
**Bad**
```rust
fn work(x: &dyn Display) { println!("{x}"); } // vtable call
```

**Good**
```rust
fn work<T: Display>(x: &T) { println!("{x}"); } // monomorphized
```
- 🦀 Clippy: `trait_object` (contextual) https://rust-lang.github.io/rust-clippy/master/  

---

### 2.7 Avoid Unnecessary `.to_string()`/`format!`
```yaml
id: rs_avoid_needless_string_alloc
langs: [rust]
category: allocation
severity: warning
```
**Bad**
```rust
let s = "value: ".to_string() + &x.to_string();
```

**Good**
```rust
use std::fmt::Write;
let mut s = String::with_capacity(32);
write!(&mut s, "value: {x}")?;
```
- 🦀 Clippy: `string_add_assign`, `inefficient_to_string`  

---

### 2.8 MaybeUninit for Large Buffers
```yaml
id: rs_maybeuninit_for_large_buffers
langs: [rust]
category: allocation
severity: advanced
```
Avoid zero-filling when immediately overwriting large buffers.
- 📖 MaybeUninit: https://doc.rust-lang.org/nomicon/uninitialized.html

---

### 2.9 Choose HashMap/BTreeMap Appropriately
```yaml
id: rs_choose_map_wisely
langs: [rust]
category: algorithm
severity: info
```
Use `HashMap` for random access; `BTreeMap` for ordered/range queries. Consider faster hashers (e.g., `ahash`) when DoS-resistance isn’t required.

---

### 2.10 Parallelism via Rayon
```yaml
id: rs_rayon_parallelism
langs: [rust]
category: concurrency
severity: info
```
**Good**
```rust
use rayon::prelude::*;
let sum: i64 = data.par_iter().map(|x| *x as i64).sum();
```
- 📖 Rayon: https://docs.rs/rayon

---

### 2.11 Async for Massive Concurrency
```yaml
id: rs_async_massive_concurrency
langs: [rust]
category: concurrency
severity: info
```
Use async executors (Tokio) for I/O-bound tasks.
- 📖 Tokio: https://tokio.rs/

---

### 2.12 LTO/PGO & Release Builds
```yaml
id: rs_enable_lto_pgo_release
langs: [rust]
category: compiler
severity: info
```
Build with `--release`, consider `-C lto` and PGO for hot paths.
- 📖 PGO: https://doc.rust-lang.org/rustc/profile-guided-optimization.html

---

## 3) Cross-Language Patterns

| Principle | Go | Rust |
|---|---|---|
| Minimize allocations | Preallocate, sync.Pool | Borrow, with_capacity |
| Contiguous data | Slices | Vec |
| Hoist invariants | `len` outside loop | iterators fuse work |
| Batch I/O/FFI | bufio, cgo batching | read/write buffers, FFI batching |
| Bound concurrency | worker pools | Rayon/threadpools/async |
| Avoid runtime dynamism | reflection | dyn Trait in hot paths |
| Measure | `pprof`, bench | criterion, flamegraph |

---

## 4) Linter Mapping Index

- Go staticcheck: SA6000 (regex in loop), SA6001 (byte→string map key), SA6002 (sync.Pool values), SA6005 (EqualFold), SA6006 (Write vs WriteString) — https://staticcheck.io/docs/checks  
- Go gocritic: hugeParam — https://go-critic.github.io/overview.html#hugeparam-ref  
- Rust Clippy: redundant_clone, ptr_arg, string_add_assign, trait_object — https://rust-lang.github.io/rust-clippy/master/

---

## 5) References (selected)

- Effective Go — https://go.dev/doc/effective_go  
- Go strings & Builder — https://go.dev/blog/strings  
- Go pipelines (concurrency) — https://go.dev/blog/pipelines  
- cgo performance — https://go.dev/doc/cgo#perf  
- staticcheck rules — https://staticcheck.io/docs/checks  
- gocritic rules — https://go-critic.github.io/overview.html  
- Rust Book — Ownership — https://doc.rust-lang.org/book/ch04-01-what-is-ownership.html  
- Rust Performance Book — https://nnethercote.github.io/perf-book/  
- Clippy lints — https://rust-lang.github.io/rust-clippy/master/  
- Rayon — https://docs.rs/rayon  
- Tokio — https://tokio.rs/  
- MaybeUninit — https://doc.rust-lang.org/nomicon/uninitialized.html  
- PGO (Rust) — https://doc.rust-lang.org/rustc/profile-guided-optimization.html
