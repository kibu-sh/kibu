package cuecore

import (
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoad(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"load": func(ts *testscript.TestScript, neg bool, args []string) {
				cwd := ts.Getenv("WORK")
				var mod struct {
					Nested struct {
						Name string
					}
				}

				_, err := LoadWithDefaults(cwd, []string{"module.cue"},
					WithValidation(),
					WithBasicDecoder(&mod),
				)
				ts.Check(err)
				require.Equal(t, "example", mod.Nested.Name)
			},
		},
	})
}
