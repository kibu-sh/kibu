package workspace

import (
	cuemod "github.com/discernhq/devx/cue.mod"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(e *testscript.Env) error {
			cwd := e.Getenv("WORK")
			return cuemod.Copy(filepath.Join(cwd, "cue.mod"))
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"load": func(ts *testscript.TestScript, neg bool, args []string) {
				config, err := LoadConfig(LoadConfigParams{
					DetermineRootParams: DetermineRootParams{
						StartDir:     args[0],
						SearchSuffix: ".devx/workspace.cue",
					},
					LoaderFunc: CueLoader,
				})
				ts.Check(err)
				require.NotNil(t, config)
				require.NotNil(t, config.ConfigStore.EncryptionKeys)
				require.Equal(t, "projects/PROJECT_ID/locations/global/keyRings/KEYRING_ID/cryptoKeys/KEY_ID", config.ConfigStore.EncryptionKeys[0].Key)
			},
			"determine": func(ts *testscript.TestScript, neg bool, args []string) {
				cwd := ts.Getenv("WORK")
				root, err := DetermineRoot(DetermineRootParams{
					StartDir:     args[0],
					SearchSuffix: ".devx",
				})
				require.Equal(t, cwd, root)
				require.FileExistsf(t, filepath.Join(cwd, ".devx/workspace.cue"), "expected to find .devx in %s", cwd)
				ts.Check(err)
			},
		},
	})
}
