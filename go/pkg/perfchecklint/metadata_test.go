package perfchecklint

import "testing"

func TestRulesExposeAllMetadata(t *testing.T) {
	rules, err := Rules()
	if err != nil {
		t.Fatalf("Rules() error: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("expected at least one rule")
	}
	for _, rule := range rules {
		if rule.ID == "" {
			t.Fatal("rule missing ID")
		}
		if rule.Code == 0 {
			t.Fatalf("rule %s missing numeric code", rule.ID)
		}
		if rule.Category == "" {
			t.Fatalf("rule %s missing category", rule.ID)
		}
		if rule.Summary == "" || rule.Fix == "" {
			t.Fatalf("rule %s missing summary/fix", rule.ID)
		}
	}

	first := rules[0]
	lookup, ok, err := LookupRule(first.ID)
	if err != nil {
		t.Fatalf("LookupRule error: %v", err)
	}
	if !ok {
		t.Fatalf("LookupRule(%s) not found", first.ID)
	}
	if lookup.Code != first.Code {
		t.Fatalf("LookupRule returned mismatched code: got %d want %d", lookup.Code, first.Code)
	}
	if len(lookup.Languages) == 0 {
		t.Fatalf("expected languages for %s", lookup.ID)
	}
	original := lookup.Languages[0]
	lookup.Languages[0] = "mutated"
	refetch, ok, err := LookupRule(first.ID)
	if err != nil || !ok {
		t.Fatalf("LookupRule second pass failed: ok=%v err=%v", ok, err)
	}
	if refetch.Languages[0] != original {
		t.Fatalf("expected lookup to be immutable; got %s", refetch.Languages[0])
	}
}
