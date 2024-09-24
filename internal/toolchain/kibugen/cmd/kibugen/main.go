package main

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibugen"
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"golang.org/x/tools/go/analysis"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pipeline.Main(pipeline.Config{
		Patterns:  []string{"./..."},
		Dir:       cwd,
		Analyzers: []*analysis.Analyzer{kibugen.Analyzer},
	})
}
