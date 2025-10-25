package perfchecklint

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestBuildDefaults(t *testing.T) {
	bundle := Build(BuildOptions{})
	if bundle.Name != "perfcheck" {
		t.Fatalf("unexpected name %q", bundle.Name)
	}
	if bundle.Description == "" {
		t.Fatal("description should default")
	}
	if len(bundle.Analyzers) == 0 {
		t.Fatal("expected analyzers")
	}
}

func TestBuildGoanalysisUsesConstructor(t *testing.T) {
	type result struct {
		name  string
		desc  string
		count int
	}

	builder := func(name, desc string, analyzers ...*analysis.Analyzer) result {
		return result{name: name, desc: desc, count: len(analyzers)}
	}

	r := BuildGoanalysis(builder, BuildOptions{Name: "custom", Description: "desc"})
	if r.name != "custom" || r.desc != "desc" {
		t.Fatalf("unexpected metadata: %+v", r)
	}
	if r.count == 0 {
		t.Fatal("expected analyzers to be forwarded")
	}
}
