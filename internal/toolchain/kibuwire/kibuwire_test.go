package kibuwire

import (
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzerPath := filepath.Join(testdata, "analyzer")
	results := analysistest.Run(t, analyzerPath,
		Analyzer, "./...")

	providers, ok := results[0].Result.(ProviderList)
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
	//srcfiles := filepath.Join(testdata, "src")

	testscript.Run(t, testscript.Params{
		Dir:      scripts,
		TestWork: true,
		Setup: func(env *testscript.Env) error {
			// inject application env
			//env.Vars = append(env.Vars, getGoEnv(t)...)
			//env.Setenv("GOWORK", "")
			//env.Setenv("GOMOD", "")
			//return os.CopyFS(env.WorkDir, os.DirFS(srcfiles))
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"kibuwire": func(ts *testscript.TestScript, neg bool, args []string) {
				root := args[0]
				genDir := args[1]
				patterns := args[2:]

				cfg := pipeline.ConfigDefaults().
					WithDir(root).
					WithPatterns(patterns).
					WithAnalyzers([]*analysis.Analyzer{Analyzer})

				results, err := pipeline.Run(cfg)
				ts.Check(err)

				_, err = SaveArtifacts(root, results)
				ts.Check(err)

				_ = exec.Command("open", root).Run()
			},
		},
	})
}
