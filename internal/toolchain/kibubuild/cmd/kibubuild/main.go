package main

import (
	"github.com/kibu-sh/kibu/internal/toolchain/kibugenv2"
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"golang.org/x/tools/go/analysis"
	"log"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	args := os.Args[1:]

	err = pipeline.Run(pipeline.Config{
		Patterns:         args,
		FactStore:        pipeline.NoOpFactStore{},
		Analyzers:        []*analysis.Analyzer{kibugenv2.Analyzer},
		RunDespiteErrors: true,
		LoaderConfig:     pipeline.PackageLoaderConfig(cwd),
	})

	if err != nil {
		log.Fatal(err)
	}

	return
}
