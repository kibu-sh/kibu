package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/discernhq/devx/cmd/devx/cmd/cliflags"
	"github.com/discernhq/devx/internal/workspace"
	"github.com/discernhq/devx/pkg/config"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type ConfigSetCmd struct {
	*cobra.Command
}

type NewConfigSetCmdParams struct {
	Store    config.Store
	Settings workspace.ConfigStoreSettings
}

func NewConfigSetCmd(params NewConfigSetCmdParams) (cmd ConfigSetCmd) {
	cmd.Command = &cobra.Command{
		Use:     "set",
		Short:   "set",
		Long:    `set`,
		PreRunE: cobra.ExactValidArgs(1),
		RunE:    newConfigSetRunE(params),
	}

	// TODO: don't ignore these
	_ = cliflags.ConfigSetFromFile.BindToCommand(cmd.Command)
	_ = cliflags.ConfigSetFromEnvFile.BindToCommand(cmd.Command)
	_ = cliflags.ConfigSetFromLiteral.BindToCommand(cmd.Command)
	return
}

func loadEnvFile(file string) (data any, err error) {
	if file, err = filepath.Abs(file); err != nil {
		return
	}
	return godotenv.Read(file)
}

func loadJSONFile(file string) (data any, err error) {
	if file, err = filepath.Abs(file); err != nil {
		return
	}

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&data)
	return
}

func fromLiterals(literals []string) (data any, err error) {
	literalMap := make(map[string]string)
	for _, literal := range literals {
		key, value, err := parseLiteral(literal)
		if err != nil {
			return nil, err
		}
		literalMap[key] = value
	}
	data = literalMap
	return
}

func parseLiteral(literal string) (key string, value string, err error) {
	parts := strings.SplitN(literal, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid literal %q", literal)
	}
	return parts[0], parts[1], nil
}

func fromStdIn() (data any, err error) {
	err = json.NewDecoder(os.Stdin).Decode(&data)
	return
}

func loadSetDataFromFlags() (data any, err error) {
	file := cliflags.ConfigSetFromEnvFile.Value()
	envFile := cliflags.ConfigSetFromEnvFile.Value()
	literals := cliflags.ConfigSetFromLiteral.Value()

	if file == "-" {
		return fromStdIn()
	}

	if file != "" {
		return loadJSONFile(file)
	}

	if envFile != "" {
		return loadEnvFile(cliflags.ConfigSetFromEnvFile.Value())
	}

	if len(literals) > 0 {
		return fromLiterals(literals)
	}

	err = errors.New("at least one of --from-file, --from-env-file, or --from-literal must be specified")
	return
}

func newConfigSetRunE(params NewConfigSetCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		key, err := params.Settings.KeyByEnv(cliflags.Environment.Value())
		if err != nil {
			return
		}

		data, err := loadSetDataFromFlags()
		if err != nil {
			return
		}

		_, err = params.Store.Set(context.Background(), config.SetParams{
			Path:          args[0],
			Data:          data,
			EncryptionKey: key,
		})

		return
	}
}
