package perfchecklint

import "golang.org/x/tools/go/analysis"

// Bundle describes the analyzer set along with human-readable metadata for
// downstream runners.
type Bundle struct {
	Name        string
	Description string
	Analyzers   []*analysis.Analyzer
}

// BuildOptions tunes the metadata attached to the exported analyzer bundle.
//
// Name defaults to "perfcheck" and Description defaults to
// "Performance-by-default analyzers" when unset.
type BuildOptions struct {
	Name        string
	Description string
}

// Build returns the perfcheck analyzer bundle with normalized metadata.
func Build(opts BuildOptions) Bundle {
	name := opts.Name
	if name == "" {
		name = "perfcheck"
	}
	desc := opts.Description
	if desc == "" {
		desc = "Performance-by-default analyzers"
	}
	return Bundle{
		Name:        name,
		Description: desc,
		Analyzers:   Analyzers(),
	}
}

// BuildGoanalysis constructs a bundle and immediately feeds it into the
// provided constructor (for example, golangci-lint's goanalysis.NewLinter).
func BuildGoanalysis[T any](
	ctor func(name, description string, analyzers ...*analysis.Analyzer) T,
	opts BuildOptions,
) T {
	bundle := Build(opts)
	return ctor(bundle.Name, bundle.Description, bundle.Analyzers...)
}
