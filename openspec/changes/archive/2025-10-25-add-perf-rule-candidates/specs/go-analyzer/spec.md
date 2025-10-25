## MODIFIED Requirements
### Requirement: Go Performance Analyzer
The system SHALL provide a go/analysis-based linter that surfaces performance-by-default violations.
#### Scenario: Detect linked-list usage
- **WHEN** Go code instantiates or imports `container/list` for general sequence storage
- **THEN** the analyzer SHALL emit a `perf_avoid_linked_list` diagnostic recommending slices instead of linked lists.

#### Scenario: Detect small-value mutexes
- **WHEN** Go code wraps a primitive value (bool, numeric counter, pointer) with `sync.Mutex` or `sync.RWMutex` and performs only single-value operations in the critical section
- **THEN** the analyzer SHALL report `perf_atomic_for_small_lock` and suggest using an atomic type.

#### Scenario: Detect defer statements inside loops
- **WHEN** Go code executes `defer` inside a loop body
- **THEN** the analyzer SHALL warn with `perf_no_defer_in_loop`, explaining that each deferred call is delayed until the surrounding function returns.

#### Scenario: Detect rune slice conversions for iteration
- **WHEN** Go code converts a string to `[]rune` solely to range over it
- **THEN** the analyzer SHALL emit `perf_avoid_rune_conversion` and advise iterating the string directly.

#### Scenario: Detect unbuffered I/O hot paths
- **WHEN** Go code performs many small writes or reads on an `io.Writer`/`io.Reader` (or `fmt.Fprint*` helpers) without `bufio`
- **THEN** the analyzer SHALL emit `perf_use_buffered_io`, recommending buffered I/O.

#### Scenario: Detect needless heap allocation of small structs
- **WHEN** Go code allocates or passes pointers to small structs/values that could remain on the stack without escaping
- **THEN** the analyzer SHALL emit `perf_prefer_stack_alloc` and explain that stack allocation avoids garbage and pointer indirection.
