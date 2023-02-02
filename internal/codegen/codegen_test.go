package codegen

import (
	"github.com/rogpeppe/go-internal/testscript"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "../", "testdata")
	testscript.Run(t, testscript.Params{
		Dir: testdata,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"parse": func(ts *testscript.TestScript, neg bool, args []string) {
				err := Generate(GenerateParams{
					Dir:       args[0],
					Pipeline:  DefaultPipeline(),
					Patterns:  []string{"./..."},
					OutputDir: filepath.Join(args[0], "gen"),
				})
				ts.Check(err)
			},
		},
	})
}
