package perfchecklint

import (
	"sync"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
)

// RuleMetadata provides a read-only view of the canonical perfcheck rule registry.
type RuleMetadata struct {
	ID          string
	Category    string
	Severity    string
	Description string
	Summary     string
	Fix         string
	Languages   []string
	Code        uint32
}

var (
	once          sync.Once
	cachedRules   []RuleMetadata
	cachedLookup  map[string]RuleMetadata
	registryError error
)

// Rules returns the immutable list of perfcheck rules sorted by identifier.
//
// Callers receive a copy so they may freely modify the slice without affecting
// other consumers.
func Rules() ([]RuleMetadata, error) {
	loadRegistry()
	if registryError != nil {
		return nil, registryError
	}
	out := make([]RuleMetadata, len(cachedRules))
	for i := range cachedRules {
		out[i] = cachedRules[i]
		out[i].Languages = append([]string(nil), out[i].Languages...)
	}
	return out, nil
}

// MustRules mirrors Rules but panics if the embedded registry cannot be loaded.
func MustRules() []RuleMetadata {
	loadRegistry()
	if registryError != nil {
		panic(registryError)
	}
	out := make([]RuleMetadata, len(cachedRules))
	for i := range cachedRules {
		out[i] = cachedRules[i]
		out[i].Languages = append([]string(nil), out[i].Languages...)
	}
	return out
}

// LookupRule returns metadata for a rule by identifier.
func LookupRule(id string) (RuleMetadata, bool, error) {
	loadRegistry()
	if registryError != nil {
		return RuleMetadata{}, false, registryError
	}
	rule, ok := cachedLookup[id]
	if !ok {
		return RuleMetadata{}, false, nil
	}
	rule.Languages = append([]string(nil), rule.Languages...)
	return rule, true, nil
}

func loadRegistry() {
	once.Do(func() {
		reg, err := ruleset.Default()
		if err != nil {
			registryError = err
			return
		}
		all := reg.All()
		cachedRules = make([]RuleMetadata, len(all))
		cachedLookup = make(map[string]RuleMetadata, len(all))
		for i := range all {
			rule := all[i]
			metadata := RuleMetadata{
				ID:          rule.ID,
				Category:    rule.Category,
				Severity:    rule.Severity,
				Description: rule.Description,
				Summary:     rule.ProblemSummary,
				Fix:         rule.FixHint,
				Languages:   append([]string(nil), rule.Langs...),
				Code:        rule.Code,
			}
			cachedRules[i] = metadata
			cachedLookup[metadata.ID] = metadata
		}
	})
}
