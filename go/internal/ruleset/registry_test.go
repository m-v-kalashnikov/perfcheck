package ruleset

import "testing"

func TestDefaultRegistryLoads(t *testing.T) {
	reg, err := Default()
	if err != nil {
		t.Fatalf("Default() unexpected error: %v", err)
	}
	if len(reg.All()) == 0 {
		t.Fatal("registry contained no rules")
	}

	seen := make(map[uint32]string, len(reg.All()))
	for _, rule := range reg.All() {
		if rule.ID == "" {
			t.Fatalf("rule returned empty id: %+v", rule)
		}
		if prev, ok := seen[rule.Code]; ok && prev != rule.ID {
			t.Fatalf("hash collision between %q and %q", prev, rule.ID)
		}
		seen[rule.Code] = rule.ID
	}
}

func TestRulesForLangReturnsCopy(t *testing.T) {
	reg := MustDefault()

	rules := reg.RulesForLang("go")
	if len(rules) == 0 {
		t.Fatal("expected go rules")
	}

	rules[0].ID = "mutated"

	fresh := reg.RulesForLang("go")
	if fresh[0].ID == "mutated" {
		t.Fatal("returned slice is not isolated copy")
	}
}
