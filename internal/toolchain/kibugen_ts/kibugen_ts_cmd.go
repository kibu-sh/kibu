package kibugen_ts

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"golang.org/x/tools/go/analysis"
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
	}

	err = os.MkdirAll(genDir, 0755)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to make -out relative to root"))
	}

	cfg := pipeline.ConfigDefaults().
		WithDir(root).
		WithPatterns(patterns).
		WithAnalyzers([]*analysis.Analyzer{Analyzer})

	results, pkgs, err := pipeline.Run(cfg)
	if err != nil {
		return 1, errors.Join(err, errors.New("failed to run pipeline"))
	}

	if len(pkgs) == 0 {
		return 1, errors.New("no packages found")
	}

	artifacts := gatherArtifacts(results)

	outputDir := filepath.Join(genDir, "kibugen_ts")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 1, errors.Join(err, fmt.Errorf("failed to create output directory %s", outputDir))
	}

	if err := unpackEmbeddedFiles(outputDir); err != nil {
		return 1, errors.Join(err, errors.New("failed to unpack embedded files"))
	}

	for _, artifact := range artifacts {
		relPath := artifact.OutputPath()
		outPath := filepath.Join(outputDir, relPath)
		outDir := filepath.Dir(outPath)

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return 1, errors.Join(err, fmt.Errorf("failed to create directory %s", outDir))
		}

		if err := os.WriteFile(outPath, []byte(artifact.Contents()), 0644); err != nil {
			return 1, errors.Join(err, fmt.Errorf("failed to write file %s", outPath))
		}

		fmt.Printf("Generated: %s\n", outPath)
	}

	return 0, nil
}

func unpackEmbeddedFiles(outputDir string) error {
	entries, err := EmbeddedFiles.ReadDir("embedded")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := EmbeddedFiles.ReadFile(filepath.Join("embedded", entry.Name()))
		if err != nil {
			return err
		}

		outPath := filepath.Join(outputDir, entry.Name())
		if err := os.WriteFile(outPath, content, 0644); err != nil {
			return err
		}

		fmt.Printf("Unpacked: %s\n", outPath)
	}

	return nil
}

func gatherArtifacts(results []*analysis.Pass) []*artifact {
	var artifacts []*artifact
	for _, pass := range results {
		for _, result := range pass.ResultOf {
			if art, ok := result.(*artifact); ok {
				artifacts = append(artifacts, art)
			}
		}
	}
	return artifacts
}
