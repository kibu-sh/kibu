package codegen

import (
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
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

func TestPackageScopedID(t *testing.T) {
	var tests = []struct {
		pkg  string
		name string
		want string
	}{
		{"go/context", "Example", "ContextExample"},
		{"other/context", "Example", "ContextExample"},
		{"github.com/testing/middleware", "Example", "MiddlewareExample"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, buildPackageScopedID(tt.pkg, tt.name))
	}
}
