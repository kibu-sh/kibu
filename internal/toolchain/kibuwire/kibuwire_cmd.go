package kibuwire

import (
	"errors"
	"flag"
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"golang.org/x/tools/go/analysis"
	"os"
	"path/filepath"
	"strings"
)

func Main() (int, error) {
	var root string
	var genDir string
	var patterns []string

	cwd, err := os.Getwd()
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to get current working directory"))
	}

	// defaults to cwd/gen
	genDir = filepath.Join(cwd, "gen")
	fset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fset.StringVar(&root, "cwd", cwd, "current working directory")
	fset.StringVar(&genDir, "out", "", "output directory")

	err = fset.Parse(os.Args[1:])
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to parse flags"))
	}

	patterns = fset.Args()

	if !filepath.IsAbs(genDir) {
		genDir = filepath.Clean(filepath.Join(root, genDir))
		if err != nil {
			return 1, errors.Join(err, errors.New("failed to make -out relative to root"))
		}
	}

	cfg := pipeline.ConfigDefaults().
		WithDir(root).
		WithPatterns(patterns).
		WithAnalyzers([]*analysis.Analyzer{Analyzer})

	results, pkgs, err := pipeline.Run(cfg)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to run pipeline"))
	}

	wireModPrefix := strings.TrimPrefix(genDir, pkgs[0].Module.Dir)
	providerArtifacts := modspecv2.GatherResults[*Artifact](results)
	wiremod := buildKibuWireModule(wireModPrefix, providerArtifacts)
	artifacts := modspecv2.GatherResults[modspecv2.Artifact](results)
	artifacts = append(artifacts, wiremod)
	_, err = modspecv2.SaveArtifacts(pkgs[0].Module, artifacts)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to save artifacts"))
	}
	return 0, nil
}
