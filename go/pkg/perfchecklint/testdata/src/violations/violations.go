package violations

import (
	"container/list"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// perf_avoid_string_concat_loop
func concat(items []string) string {
	out := ""
	for _, item := range items {
		out += item // want "[perf_avoid_string_concat_loop]"
	}
	return out
}

// perf_regex_compile_once
func regexCount(inputs []string, expr string) int {
	count := 0
	for _, in := range inputs {
		if regexp.MustCompile(expr).MatchString(in) { // want "[perf_regex_compile_once]"
			count++
		}
	}
	return count
}

// perf_preallocate_collections
func collect(numbers []int) []int {
	var out []int
	for _, n := range numbers {
		out = append(out, n) // want "[perf_preallocate_collections]"
	}
	return out
}

// perf_avoid_reflection_dynamic
func kinds(values []any) []reflect.Kind {
	var result []reflect.Kind
	for _, v := range values {
		result = append(result, reflect.TypeOf(v).Kind()) // want "[perf_avoid_reflection_dynamic]"
	}
	return result
}

// perf_bound_concurrency
func spawnAll(tasks []func()) {
	for _, task := range tasks {
		go task() // want "[perf_bound_concurrency]"
	}
}

// perf_equal_fold_compare
func equalInsensitive(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b) // want "[perf_equal_fold_compare]"
}

// perf_syncpool_store_pointers
type pooled struct{ buf [64]byte }

func store(pool *sync.Pool, value pooled) {
	pool.Put(value) // want "[perf_syncpool_store_pointers]"
}

// perf_writer_prefer_bytes
func writeBytes(w io.Writer, payload []byte) (int, error) {
	return io.WriteString(w, string(payload)) // want "[perf_writer_prefer_bytes]"
}

// perf_avoid_linked_list
func linked(items []int) *list.List {
	ll := list.New()
	for _, item := range items {
		ll.PushBack(item) // want "[perf_avoid_linked_list]"
	}
	return ll
}

// perf_atomic_for_small_lock
type counter struct {
	mu  sync.Mutex
	val int
}

func (c *counter) set(v int) {
	c.mu.Lock()
	c.val = v // want "[perf_atomic_for_small_lock]"
	c.mu.Unlock()
}

// perf_no_defer_in_loop
func closeLater(files []io.Closer) {
	for _, f := range files {
		defer f.Close() // want "[perf_no_defer_in_loop]"
	}
}

// perf_avoid_rune_conversion
func countRunes(s string) int {
	count := 0
	for range []rune(s) { // want "[perf_avoid_rune_conversion]"
		count++
	}
	return count
}

// perf_use_buffered_io
func writeLoop(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil { // want "[perf_use_buffered_io]"
			return err
		}
	}
	return nil
}

// perf_prefer_stack_alloc
type point struct {
	x, y int
}

func newPoint(x, y int) *point {
	return &point{x: x, y: y} // want "[perf_prefer_stack_alloc]"
}

// helper to keep package referenced
func use(values ...any) {
	fmt.Fprint(io.Discard, values...)
}
