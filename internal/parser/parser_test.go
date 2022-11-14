package parser

import (
	"github.com/rogpeppe/go-internal/testscript"
	"testing"
)

func TestParser(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"parse": func(ts *testscript.TestScript, neg bool, args []string) {
				// _, err := ParseDir(args[0])
				_, err := experimentalParse(args[0])
				ts.Check(err)
			},
		},
	})
}
