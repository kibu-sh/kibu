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
			"kibugenv2": func(ts *testscript.TestScript, neg bool, args []string) {
				cfg := pipeline.ConfigDefaults().
					WithPatterns(args).
					WithDir(args[0]).
					WithPatterns(args[1:]).
					WithAnalyzers([]*analysis.Analyzer{Analyzer})

				ts.Check(pipeline.Run(cfg))
			},
		},
	})
}

//func getGoEnv(t *testing.T) []string {
//	t.Helper()
//	cmd := exec.Command("go", "env")
//	out, err := cmd.Output()
//	if err != nil {
//		t.Fatal(err)
//	}
//	return strings.Split(string(out), "\n")
//}
