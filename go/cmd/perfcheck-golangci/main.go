package main

import (
	"github.com/yourname/perfcheck/go/internal/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analyzer.All()...)
}
