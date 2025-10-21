package ruleset

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:generate go run ../tools/copyrules -src ../../../perfcheck-core/config/default_rules.tsv -dst data/default_rules.tsv

// defaultRulesBundle embeds the canonical rule set so the analyzer can run offline.
//
//go:embed data/default_rules.tsv
var defaultRulesBundle []byte

var (
	loadOnce sync.Once
	cached   *Registry
	loadErr  error
)

// Rule stores normalized rule metadata for fast lookups at analysis time.
type Rule struct {
	ID          string
	Langs       []string
	Description string
	Category    string
	Severity    string
	Code        uint32
}

// Registry groups rules by language and identifier for efficient querying.
type Registry struct {
	byLang map[string][]Rule
	byID   map[string]Rule
	all    []Rule
}

// Default returns the singleton registry hydrated from the embedded rule bundle.
func Default() (*Registry, error) {
	loadOnce.Do(func() {
		if len(defaultRulesBundle) == 0 {
			loadErr = errors.New("ruleset: default rule bundle is empty")
			return
		}

		rules, err := parseTSV(defaultRulesBundle)
		if err != nil {
			loadErr = err
			return
		}

		if len(rules) == 0 {
			loadErr = errors.New("ruleset: no rules defined in bundle")
			return
		}

		cached = buildRegistry(rules)
	})

	return cached, loadErr
}

// MustDefault loads the registry or panics if hydration fails.
func MustDefault() *Registry {
	reg, err := Default()
	if err != nil {
		panic(err)
	}
	return reg
}

func buildRegistry(rules []Rule) *Registry {
	byID := make(map[string]Rule, len(rules))
	byLang := make(map[string][]Rule, 4)
	all := make([]Rule, 0, len(rules))

	for _, rule := range rules {
		if rule.ID == "" {
			continue
		}

		rule.Code = hash(rule.ID)
		rule.Severity = strings.ToLower(rule.Severity)
		rule.Category = strings.ToLower(rule.Category)

		normalizedLangs := make([]string, 0, len(rule.Langs))
		for _, lang := range rule.Langs {
			lang = strings.ToLower(strings.TrimSpace(lang))
			if lang == "" {
				continue
			}
			normalizedLangs = append(normalizedLangs, lang)
			byLang[lang] = append(byLang[lang], rule)
		}
		rule.Langs = normalizedLangs

		byID[rule.ID] = rule
		all = append(all, rule)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	for lang, rules := range byLang {
		sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
		byLang[lang] = rules
	}

	return &Registry{
		byLang: byLang,
		byID:   byID,
		all:    all,
	}
}

// RuleByID returns the rule metadata by identifier.
func (r *Registry) RuleByID(id string) (Rule, bool) {
	rule, ok := r.byID[id]
	return rule, ok
}

// RulesForLang returns a copy of the rules matching the provided language token.
func (r *Registry) RulesForLang(lang string) []Rule {
	lang = strings.ToLower(lang)
	rules := r.byLang[lang]
	if len(rules) == 0 {
		return nil
	}
	out := make([]Rule, len(rules))
	copy(out, rules)
	return out
}

// All returns an immutable copy of the full rule list.
func (r *Registry) All() []Rule {
	out := make([]Rule, len(r.all))
	copy(out, r.all)
	return out
}

func hash(id string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return h.Sum32()
}

func parseTSV(data []byte) ([]Rule, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	rules := make([]Rule, 0, 16)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if lineNum == 1 || line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			return nil, errors.New("ruleset: invalid field count on line " + strconv.Itoa(lineNum))
		}

		langs := parseLangs(fields[1])

		rules = append(rules, Rule{
			ID:          strings.TrimSpace(fields[0]),
			Langs:       langs,
			Description: strings.TrimSpace(fields[2]),
			Category:    strings.TrimSpace(fields[3]),
			Severity:    strings.TrimSpace(fields[4]),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func parseLangs(raw string) []string {
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
