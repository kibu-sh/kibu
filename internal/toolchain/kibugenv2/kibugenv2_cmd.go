package kibugenv2

import (
	"errors"
	"flag"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"golang.org/x/tools/go/analysis"
	"os"
)

func Main() (int, error) {
	var root string
	var patterns []string

	cwd, err := os.Getwd()
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to get current working directory"))
	}

	fset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fset.StringVar(&root, "cwd", cwd, "current working directory")

	if err = fset.Parse(os.Args[1:]); err != nil {
		return 1, errors.Join(err, errors.New("failed to parse flags"))
	}

	cfg := pipeline.ConfigDefaults().
		WithDir(root).
		WithPatterns(patterns).
		WithAnalyzers([]*analysis.Analyzer{Analyzer})

	results, pkgs, err := pipeline.Run(cfg)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to run pipeline"))
	}

	artifacts := modspecv2.GatherResults[modspecv2.Artifact](results)
	_, err = modspecv2.SaveArtifacts(pkgs[0].Module, artifacts)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to save artifacts"))
	}

	return 0, nil
}
