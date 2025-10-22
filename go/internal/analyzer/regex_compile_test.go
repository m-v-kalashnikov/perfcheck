package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/ruleset"
	"golang.org/x/tools/go/analysis"
)

func TestRegexCompileLoopAnalyzer(t *testing.T) {
	src := `package sample

import "regexp"

func matchAll(items []string) int {
	count := 0
	for _, item := range items {
		re := regexp.MustCompile("^foo$")
		if re.MatchString(item) {
			count++
		}
	}
	return count
}
`
	diags := runRegex(src)
	require.Len(t, diags, 1)
}

func runRegex(src string) []analysis.Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "regex.go", src, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("sample", fset, []*ast.File{file}, info); err != nil {
		panic(err)
	}

	var diags []analysis.Diagnostic
	rule, ok := ruleset.MustDefault().RuleByID("perf_regex_compile_once")
	if !ok {
		panic("rule not found")
	}

	pass := analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}

	var fnBody *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			fnBody = fn.Body
			break
		}
	}
	if fnBody == nil {
		panic("no function body found")
	}
	checkRegexCalls(&pass, fnBody, rule)
	return diags
}
