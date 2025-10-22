package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/analyzer"
)

func main() {
	multichecker.Main(analyzer.All()...)
}
