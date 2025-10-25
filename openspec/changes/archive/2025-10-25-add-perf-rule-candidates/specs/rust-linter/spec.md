## MODIFIED Requirements
### Requirement: Rust Performance Linter
The system SHALL provide a Rust linter capable of scanning crate source files for performance-by-default violations.
#### Scenario: Detect linked-list usage
- **WHEN** Rust code imports or instantiates `std::collections::LinkedList` for general-purpose sequences
- **THEN** the linter SHALL emit `perf_avoid_linked_list`, recommending `Vec`/`VecDeque` alternatives.

#### Scenario: Detect large enum variants
- **WHEN** a Rust enum variant holds data significantly larger than the others (e.g., >5× larger or an oversized array)
- **THEN** the linter SHALL emit `perf_large_enum_variant` suggesting boxing the large payload.

#### Scenario: Detect unnecessary Arc usage
- **WHEN** code wraps non-`Send + Sync` data in `Arc<T>` without evidence of multi-threaded sharing
- **THEN** the linter SHALL emit `perf_unnecessary_arc` and recommend `Rc<T>` or borrowing instead.

#### Scenario: Detect mutex-protected primitives
- **WHEN** Rust code uses `std::sync::Mutex` to guard a primitive (bool, counter, pointer) without multi-step invariants
- **THEN** the linter SHALL emit `perf_atomic_for_small_lock` recommending the matching atomic type.

#### Scenario: Detect needless iterator collect
- **WHEN** Rust code calls `.collect()` only to perform a cheap derived computation (length, iteration, immediate consumption)
- **THEN** the linter SHALL emit `perf_needless_collect`, pointing to iterator adapters that avoid allocation.

#### Scenario: Detect unnecessary heap indirection
- **WHEN** Rust code allocates a small `Copy` or <=2-word struct using `Box`, `Arc`, or `Rc` when stack storage suffices
- **THEN** the linter SHALL emit `perf_prefer_stack_alloc` to encourage stack allocation.
