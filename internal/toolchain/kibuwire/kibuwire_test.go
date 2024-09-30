package kibuwire

import (
	"github.com/kibu-sh/kibu/internal/toolchain/modspecv2"
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"path/filepath"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzerPath := filepath.Join(testdata, "analyzer")
	results := analysistest.Run(t, analyzerPath,
		Analyzer, "./...")

	artifact, ok := results[0].Result.(*Artifact)
	providers := artifact.Providers
	require.True(t, ok)
	require.NotNil(t, providers)
	require.Equal(t, 3, providers.Len())

	grouped := providers.GroupBy(GroupByFQN())
	require.Equal(t, grouped.Len(), 1, "should have 1 group")
	httpHandlers, ok := grouped.Get("github.com/kibu-sh/kibu/pkg/transport/httpx.HandlerFactory")
	require.True(t, ok)
	require.NotNil(t, httpHandlers)
	require.Equal(t, httpHandlers.Len(), 1)

	require.Equal(t, httpHandlers[0].Group.Name, "HandlerFactory")
	require.Equal(t, httpHandlers[0].Group.Import, "github.com/kibu-sh/kibu/pkg/transport/httpx")
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
			"kibuwire": func(ts *testscript.TestScript, neg bool, args []string) {
				root := args[0]
				//genDir := args[1]
				patterns := args[2:]

				cfg := pipeline.ConfigDefaults().
					WithDir(root).
					WithPatterns(patterns).
					WithAnalyzers([]*analysis.Analyzer{Analyzer})

				results, err := pipeline.Run(cfg)
				ts.Check(err)

				artifacts := modspecv2.GatherResults[modspecv2.Artifact](results)
				_, err = modspecv2.SaveArtifacts(root, artifacts)
				ts.Check(err)
			},
		},
	})
}
