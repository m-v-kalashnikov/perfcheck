package perfchecklint

import "testing"

func TestStackAllocAnalyzerFlagsSmallStructPointer(t *testing.T) {
	src := `package sample

type point struct {
	x, y int
}

func newPoint(x, y int) *point {
	return &point{x: x, y: y}
}
`

	diags := runAnalyzerOnSource(t, stackAllocAnalyzer, "stack_alloc.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_prefer_stack_alloc") {
		t.Fatalf("missing perf_prefer_stack_alloc diagnostic")
	}
}

func TestStackAllocAnalyzerIgnoresLargeStruct(t *testing.T) {
	src := `package sample

type large struct {
	buf [256]byte
}

func makeLarge() *large {
	return &large{}
}
`

	diags := runAnalyzerOnSource(t, stackAllocAnalyzer, "stack_alloc_large.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
