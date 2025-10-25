package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func TestViolationsFixtureTriggersAllRules(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	fixturePath := filepath.Join(filepath.Dir(filename), "testdata", "src", "violations", "violations.go")
	src, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fixturePath, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("violations", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatalf("type-check fixture: %v", err)
	}

	expected := map[string]bool{
		"perf_avoid_string_concat_loop": false,
		"perf_regex_compile_once":       false,
		"perf_preallocate_collections":  false,
		"perf_avoid_reflection_dynamic": false,
		"perf_bound_concurrency":        false,
		"perf_equal_fold_compare":       false,
		"perf_syncpool_store_pointers":  false,
		"perf_writer_prefer_bytes":      false,
		"perf_avoid_linked_list":        false,
		"perf_atomic_for_small_lock":    false,
		"perf_no_defer_in_loop":         false,
		"perf_avoid_rune_conversion":    false,
		"perf_use_buffered_io":          false,
		"perf_prefer_stack_alloc":       false,
	}

	for _, analyzer := range All() {
		diags := runAnalyzerOnFixture(t, analyzer, fset, file, info, pkg)
		if len(diags) == 0 {
			t.Errorf("%s: expected diagnostics, got none", analyzer.Name)
			continue
		}
		for _, diag := range diags {
			id := extractRuleID(diag.Message)
			if id == "" {
				t.Errorf("%s: unable to extract rule id from message %q", analyzer.Name, diag.Message)
				continue
			}
			if _, ok := expected[id]; ok {
				expected[id] = true
			}
		}
	}

	for id, seen := range expected {
		if !seen {
			t.Errorf("expected rule %s to trigger on fixture, but it did not", id)
		}
	}
}

func runAnalyzerOnFixture(
	t *testing.T,
	analyzer *analysis.Analyzer,
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
) []analysis.Diagnostic {
	t.Helper()

	var diags []analysis.Diagnostic
	pass := analysis.Pass{
		Analyzer:   analyzer,
		Fset:       fset,
		Files:      []*ast.File{file},
		Pkg:        pkg,
		TypesInfo:  info,
		TypesSizes: types.SizesFor("gc", runtime.GOARCH),
		ResultOf:   make(map[*analysis.Analyzer]any),
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}

	for _, req := range analyzer.Requires {
		if req == inspect.Analyzer {
			pass.ResultOf[inspect.Analyzer] = inspector.New(pass.Files)
			continue
		}
		t.Fatalf("unsupported dependency %q for analyzer %s", req.Name, analyzer.Name)
	}

	if analyzer.Run == nil {
		t.Fatalf("analyzer %s has no Run function", analyzer.Name)
	}
	if _, err := analyzer.Run(&pass); err != nil {
		t.Fatalf("analyzer %s failed: %v", analyzer.Name, err)
	}
	return diags
}

func extractRuleID(message string) string {
	if !strings.HasPrefix(message, "[") {
		return ""
	}
	if idx := strings.IndexByte(message, ']'); idx > 1 {
		return message[1:idx]
	}
	return ""
}

func runAnalyzerOnSource(t *testing.T, analyzer *analysis.Analyzer, filename, src string) []analysis.Diagnostic {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("sample", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatalf("type-check %s: %v", filename, err)
	}

	pass := analysis.Pass{
		Analyzer:   analyzer,
		Fset:       fset,
		Files:      []*ast.File{file},
		Pkg:        pkg,
		TypesInfo:  info,
		TypesSizes: types.SizesFor("gc", runtime.GOARCH),
		ResultOf:   make(map[*analysis.Analyzer]any),
	}

	var diags []analysis.Diagnostic
	pass.Report = func(d analysis.Diagnostic) {
		diags = append(diags, d)
	}

	for _, req := range analyzer.Requires {
		if req == inspect.Analyzer {
			pass.ResultOf[inspect.Analyzer] = inspector.New(pass.Files)
			continue
		}
		t.Fatalf("unsupported dependency %q for analyzer %s", req.Name, analyzer.Name)
	}

	if analyzer.Run == nil {
		t.Fatalf("analyzer %s has no Run function", analyzer.Name)
	}
	if _, err := analyzer.Run(&pass); err != nil {
		t.Fatalf("analyzer %s failed: %v", analyzer.Name, err)
	}

	return diags
}
