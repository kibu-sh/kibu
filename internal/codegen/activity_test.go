package codegen

import (
	_ "embed"
)

// func TestGenerateActivity(t *testing.T) {
// 	cwd, err := os.Getwd()
// 	require.NoError(t, err)
// 	testdata := filepath.Join(cwd, "../testdata")
// 	workflowFile := filepath.Join(testdata, "module.cue")
//
// 	mod, err := cuecore.Load(cuecore.LoadOptions{
// 		Dir:        testdata,
// 		Entrypoint: []string{workflowFile},
// 	})
// 	require.NoError(t, err)
//
// 	data, err := generateActivity(mod)
// 	require.NoError(t, err)
// 	require.NotNil(t, data)
// }
