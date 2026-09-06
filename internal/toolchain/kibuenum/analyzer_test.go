package kibuenum

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kibu-sh/kibu/internal/toolchain/pipeline"
	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/tools/go/analysis"
)

func TestAnalyzer(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
		Setup: func(env *testscript.Env) error {
			// Each archive is a real Go module using the actual SDK in this
			// checkout. No synthetic importer, SDK stub, or network is needed.
			root := filepath.Join(env.WorkDir, "src")
			if err := os.MkdirAll(root, 0755); err != nil {
				return err
			}
			mod := fmt.Sprintf("module example.test/enums\n\ngo 1.25\n\nrequire github.com/kibu-sh/kibu v0.0.0\nreplace github.com/kibu-sh/kibu => %q\n", repo)
			return os.WriteFile(filepath.Join(root, "go.mod"), []byte(mod), 0644)
		},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"kibuenum": runScriptAnalyzer,
		},
	})
}

type diagnosticStore struct {
	pipeline.NoOpFactStore
	diagnostics []analysis.Diagnostic
}

func (s *diagnosticStore) Report(d analysis.Diagnostic) {
	s.diagnostics = append(s.diagnostics, d)
}

// runScriptAnalyzer follows the sibling generator commands: load the fixture
// module through pipeline.Run, then save artifacts for testscript cmp/grep.
// The artifacts here are the in-memory spec and diagnostics, not generated code.
func runScriptAnalyzer(ts *testscript.TestScript, neg bool, args []string) {
	reverseFiles := len(args) > 0 && args[0] == "-reverse-files"
	if reverseFiles {
		args = args[1:]
	}
	if len(args) < 2 {
		ts.Fatalf("usage: kibuenum [-reverse-files] root pattern...")
	}
	root := ts.MkAbs(args[0])
	store := &diagnosticStore{}
	cfg := pipeline.ConfigDefaults().WithDir(root).WithPatterns(args[1:]).
		WithAnalyzers([]*analysis.Analyzer{Analyzer}).WithFactStore(store).
		WithRunDespiteErrors(false)
	cfg.LoaderConfig.Env = append(cfg.LoaderConfig.Env, "GOWORK=off", "GOPROXY=off")
	passes, _, err := pipeline.Run(cfg)
	if err != nil {
		fmt.Fprintln(ts.Stderr(), err)
		if !neg {
			ts.Fatalf("analyzer failed: %v", err)
		}
		return
	}
	if len(passes) == 0 {
		ts.Fatalf("no packages analyzed")
	}
	if reverseFiles {
		// Rebuild inspect.Analyzer too: reversing pass.Files alone leaves the
		// cached inspector in its original traversal order.
		store.diagnostics = nil
		runner := pipeline.NewRunner(Analyzer)
		for _, pass := range passes {
			slices.Reverse(pass.Files)
			ts.Check(runner.Execute(pass))
		}
	}
	writeScriptResults(ts, root, passes, store.diagnostics)
	if failed := len(store.diagnostics) != 0; failed != neg {
		ts.Fatalf("analyzer diagnostics=%d, expected failure=%v; see diagnostics.txt", len(store.diagnostics), neg)
	}
}
