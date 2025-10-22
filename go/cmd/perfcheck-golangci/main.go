package main

import (
	"github.com/m-v-kalashnikov/perfcheck/go/internal/analyzer"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analyzer.All()...)
}
