package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/m-v-kalashnikov/perfcheck/go/internal/analyzer"
)

func main() {
	unitchecker.Main(analyzer.All()...)
}
