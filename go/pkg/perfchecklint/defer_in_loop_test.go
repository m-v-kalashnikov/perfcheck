package perfchecklint

import "testing"

func TestDeferInLoopAnalyzerFlagsDefer(t *testing.T) {
	src := `package sample

func readAll(close func()) {
	for i := 0; i < 10; i++ {
		defer close()
	}
}
`

	diags := runAnalyzerOnSource(t, deferInLoopAnalyzer, "defer_loop.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_no_defer_in_loop") {
		t.Fatalf("missing perf_no_defer_in_loop diagnostic")
	}
}

func TestDeferInLoopAnalyzerAllowsFunctionExitDefer(t *testing.T) {
	src := `package sample

func readAll(close func()) {
	defer close()
	for range [3]int{} {
		// work
	}
}
`

	diags := runAnalyzerOnSource(t, deferInLoopAnalyzer, "defer_ok.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
