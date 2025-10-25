package perfchecklint

import "testing"

func TestAtomicSmallLockAnalyzerFlagsPrimitiveMutex(t *testing.T) {
	src := `package sample

import "sync"

type counter struct {
	mu sync.Mutex
	val int
}

func (c *counter) set(v int) {
	c.mu.Lock()
	c.val = v
	c.mu.Unlock()
}
`

	diags := runAnalyzerOnSource(t, atomicSmallLockAnalyzer, "atomic_mutex.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_atomic_for_small_lock") {
		t.Fatalf("missing perf_atomic_for_small_lock diagnostic")
	}
}

func TestAtomicSmallLockAnalyzerAllowsComplexSections(t *testing.T) {
	src := `package sample

import "sync"

type counter struct {
	mu sync.Mutex
	val int
}

func (c *counter) set(v int) {
	c.mu.Lock()
	c.val += v
	c.log()
	c.mu.Unlock()
}

func (c *counter) log() {}
`

	diags := runAnalyzerOnSource(t, atomicSmallLockAnalyzer, "atomic_mutex_ok.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
