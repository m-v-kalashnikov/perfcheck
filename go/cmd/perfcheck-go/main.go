package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/m-v-kalashnikov/perfcheck/go/pkg/perfchecklint"
)

func main() {
	unitchecker.Main(perfchecklint.Analyzers()...)
}
