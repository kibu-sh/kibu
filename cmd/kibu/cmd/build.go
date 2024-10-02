package cmd

import (
	"github.com/kibu-sh/kibu/internal/codegen"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type BuildCmd struct {
	*cobra.Command
}

type NewBuildCmdParams struct {
	loadConfig configLoaderFunc
}

func NewBuildCmd(params NewBuildCmdParams) (cmd BuildCmd) {
	cmd.Command = &cobra.Command{
		Use:   "build",
		Short: "build code",
		Long:  `build code`,
		RunE:  newBuildRunE(params),
	}
	return
}

func newBuildRunE(params NewBuildCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		config, err := params.loadConfig()
		if err != nil {
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			return
		}

		err = codegen.Generate(codegen.GenerateParams{
			Dir:       cwd,
			Patterns:  args,
			Pipeline:  codegen.DefaultPipeline(),
			OutputDir: filepath.Join(config.Root(), config.CodeGen.OutputDir),
		})
		return
	}
}
