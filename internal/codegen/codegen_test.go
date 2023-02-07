package codegen

import (
	"github.com/rogpeppe/go-internal/testscript"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate(t *testing.T) {
	cwd, _ := os.Getwd()
	// Run the testscript 10 times to ensure that the test is deterministic.
	for i := 0; i < 10; i++ {
		testscript.Run(t, testscript.Params{
			Dir: filepath.Join(cwd, "testdata"),
			Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
				"parse": func(ts *testscript.TestScript, neg bool, args []string) {
					ts.Check(Generate(GenerateParams{
						Patterns:  []string{"./..."},
						Pipeline:  DefaultPipeline(),
						Dir:       filepath.Join(args[0], "src"),
						OutputDir: filepath.Join(args[0], "gen"),
					}))
				},
			},
		})
	}
}
