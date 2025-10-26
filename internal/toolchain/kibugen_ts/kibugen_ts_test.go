package kibugen_ts

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzerPath := filepath.Join(testdata, "analyzer")
	results := analysistest.Run(t, analyzerPath,
		Analyzer, "./...")

	artifact, ok := results[0].Result.(Artifact)
	require.True(t, ok)
	require.NotNil(t, artifact)
	require.NotEmpty(t, artifact.Contents())
	require.Contains(t, artifact.Contents(), "export type")
	require.Contains(t, artifact.Contents(), "import type { HTTPClient }")
	require.Contains(t, artifact.OutputPath(), ".gen.ts")

	t.Logf("Generated TypeScript:\n%s", artifact.Contents())
}

func ResolveDir(t *testing.T, rel string) string {
	t.Helper()
	abs, err := filepath.Abs(rel)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestGenerator(t *testing.T) {
	testdata := ResolveDir(t, "testdata")
	scripts := filepath.Join(testdata, "scripts")

	testscript.Run(t, testscript.Params{
		Dir:      scripts,
		TestWork: true,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"kibugen_ts": func(ts *testscript.TestScript, neg bool, args []string) {
				var root string
				var genDir string
				var patterns []string

				fset := flag.NewFlagSet("kibugen_ts", flag.ExitOnError)
				fset.StringVar(&root, "cwd", "", "current working directory")
				fset.StringVar(&genDir, "out", "", "output directory")

				err := fset.Parse(args)
				ts.Check(err)

				patterns = fset.Args()

				if !filepath.IsAbs(genDir) {
					genDir = filepath.Clean(filepath.Join(root, genDir))
				}

				cfg := pipeline.ConfigDefaults().
					WithDir(root).
					WithPatterns(patterns).
					WithAnalyzers([]*analysis.Analyzer{Analyzer})

				results, pkgs, err := pipeline.Run(cfg)
				ts.Check(err)

				if len(pkgs) == 0 {
					ts.Fatalf("no packages found")
				}

				moduleDir := pkgs[0].Module.Dir
				artifacts := gatherArtifacts(results)

				for _, artifact := range artifacts {
					outPath := filepath.Join(moduleDir, artifact.OutputPath())
					outDir := filepath.Dir(outPath)

					if err := os.MkdirAll(outDir, 0755); err != nil {
						ts.Fatalf("failed to create directory %s: %v", outDir, err)
					}

					if err := os.WriteFile(outPath, []byte(artifact.Contents()), 0644); err != nil {
						ts.Fatalf("failed to write file %s: %v", outPath, err)
					}

					fmt.Printf("Generated: %s\n", outPath)
				}
			},
		},
	})
}
