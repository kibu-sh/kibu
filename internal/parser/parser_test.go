package parser

import (
	"github.com/rogpeppe/go-internal/testscript"
	"os"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "../", "testdata")
	testscript.Run(t, testscript.Params{
		Dir: testdata,
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"parse": func(ts *testscript.TestScript, neg bool, args []string) {
				_, err := ExperimentalParse(ExperimentalParseOpts{
					Dir:      args[0],
					Patterns: []string{"./..."},
				})
				ts.Check(err)
			},
		},
	})
}
