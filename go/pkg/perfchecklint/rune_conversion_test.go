package perfchecklint

import "testing"

func TestRuneConversionAnalyzerFlagsRange(t *testing.T) {
	src := `package sample

func count(s string) int {
	for _, r := range []rune(s) {
		if r == 'a' {
			return 1
		}
	}
	return 0
}
`

	diags := runAnalyzerOnSource(t, runeConversionAnalyzer, "runes.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_avoid_rune_conversion") {
		t.Fatalf("missing perf_avoid_rune_conversion diagnostic")
	}
}

func TestRuneConversionAnalyzerAllowsRangeString(t *testing.T) {
	src := `package sample

func count(s string) int {
	for range s {
	}
	return 0
}
`

	diags := runAnalyzerOnSource(t, runeConversionAnalyzer, "runes_ok.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
