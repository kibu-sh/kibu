package kibugenv2

import (
	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/tools/go/analysis"
	"path/filepath"
	"testing"
)

//TODO: bring this back when we're more stable
//func TestMain(m *testing.M) {
//	testscript.RunMain(m, map[string]func() int{
//		"kibugenv2": Main,
//	})
//}

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
		Dir: scripts,
		//TestWork: true,
		Setup: func(env *testscript.Env) error {
			// inject application env
			//env.Vars = append(env.Vars, getGoEnv(t)...)
			//env.Setenv("GOWORK", "")
			//env.Setenv("GOMOD", "")
			//return os.CopyFS(env.WorkDir, os.DirFS(srcfiles))
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"kibugenv2": func(ts *testscript.TestScript, neg bool, args []string) {
				root := args[0]

				cfg := pipeline.ConfigDefaults().
					WithPatterns(args).
					WithDir(root).
					WithPatterns(args[1:]).
					WithAnalyzers([]*analysis.Analyzer{Analyzer})

				results, err := pipeline.Run(cfg)
				ts.Check(err)

				_, err = SaveArtifacts(root, results)
				ts.Check(err)
			},
		},
	})
}
