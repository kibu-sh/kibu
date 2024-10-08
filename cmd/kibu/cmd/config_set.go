package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kibu-sh/kibu/cmd/kibu/cmd/cliflags"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/kibu-sh/kibu/pkg/workspace"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type ConfigSetCmd struct {
	*cobra.Command
}

type storeSettingsLoaderFunc func() (workspace.ConfigStoreSettings, error)

func provideStoreSettingsLoader(configLoader configLoaderFunc) storeSettingsLoaderFunc {
	return func() (workspace.ConfigStoreSettings, error) {
		workspaceConfig, err := configLoader()
		if err != nil {
			return workspace.ConfigStoreSettings{}, err
		}
		return workspaceConfig.ConfigStore, nil
	}
}

type NewConfigSetCmdParams struct {
	storeLoader    storeLoaderFunc
	settingsLoader storeSettingsLoaderFunc
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

type joinSecretEnvParams struct {
	Env  string
	Path string
}

func joinSecretEnvPath(params joinSecretEnvParams) string {
	return filepath.Join(params.Env, params.Path)
}

func newConfigSetRunE(params NewConfigSetCmdParams) RunE {
	return func(cmd *cobra.Command, args []string) (err error) {
		store, err := params.storeLoader()
		if err != nil {
			return
		}

		settings, err := params.settingsLoader()
		if err != nil {
			return
		}

		key, err := settings.KeyByEnv(cliflags.Environment.Value())
		if err != nil {
			return
		}

		data, err := loadSetDataFromFlags()
		if err != nil {
			return
		}

		path := joinSecretEnvPath(joinSecretEnvParams{
			Env:  cliflags.Environment.Value(),
			Path: args[0],
		})

		_, err = store.Set(context.Background(), config.SetParams{
			Path:          path,
			Data:          data,
			EncryptionKey: key,
		})

		return
	}
}
