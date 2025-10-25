package analyzer

import "testing"

func TestLinkedListAnalyzerFlagsUsage(t *testing.T) {
	src := `package sample

import "container/list"

func build(items []int) *list.List {
	ll := list.New()
	for _, item := range items {
		ll.PushBack(item)
	}
	return ll
}
`

	diags := runAnalyzerOnSource(t, linkedListAnalyzer, "linked_list.go", src)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !containsRule(diags, "perf_avoid_linked_list") {
		t.Fatalf("missing perf_avoid_linked_list diagnostic")
	}
}

func TestLinkedListAnalyzerIgnoresSlices(t *testing.T) {
	src := `package sample

func build(items []int) []int {
	out := make([]int, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
`

	diags := runAnalyzerOnSource(t, linkedListAnalyzer, "linked_list_ok.go", src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(diags))
	}
}
