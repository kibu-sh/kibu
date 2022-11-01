package build

import (
	cuemod "github.com/discernhq/devx/cue.mod"
	"github.com/rogpeppe/go-internal/testscript"
	"os"
	"path/filepath"
	"testing"
)

func TestBuild(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		//TestWork: true,
		Setup: func(env *testscript.Env) error {
			env.Setenv("HOME", os.Getenv("HOME"))
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"build": func(ts *testscript.TestScript, neg bool, args []string) {
				cwd := ts.Getenv("WORK")
				err := cuemod.Copy(filepath.Join(cwd, "cue.mod"))
				ts.Check(err)

				b, err := NewWithDefaults(cwd)
				ts.Check(err)

				err = b.Exec()
				ts.Check(err)
			},
		},
	})
}
