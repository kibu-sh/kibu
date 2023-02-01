package workspace

import (
	"context"
	"encoding/json"
	"github.com/discernhq/devx/internal/cuecore"
	"github.com/discernhq/devx/pkg/config"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"os"
	"path/filepath"
)

// FileSystemSettings configures the workspace file system observer
type FileSystemSettings struct {
	Ignore []string
}

// RemoteCacheSettings configures the workspace remote cache location
type RemoteCacheSettings struct {
	URL string
}

// ConfigStoreSettings allows user to set Vault address that's not reliant on an env var
type ConfigStoreSettings struct {
	EncryptionKeys []config.EncryptionKey
}

func (s ConfigStoreSettings) KeyByEnv(env string) (config.EncryptionKey, error) {
	for _, key := range s.EncryptionKeys {
		if key.Env == env {
			return key, nil
		}
	}

	return config.EncryptionKey{}, errors.Errorf("no encryption key found for env %s, found (%s)", env, lo.Map(s.EncryptionKeys, func(k config.EncryptionKey, _ int) string {
		return k.Env
	}))
}

// Config holds data for configuring a workspace
type Config struct {
	file                 string
	ConfigStore          ConfigStoreSettings
	FileSystem           FileSystemSettings
	RemoteCache          RemoteCacheSettings
	VersionCheckDisabled bool
}

type DetermineRootParams struct {
	StartDir     string
	SearchSuffix string
}

// DetermineRoot recursively searches the current directory and all its parents for DetermineRootParams.SearchSuffix
func DetermineRoot(params DetermineRootParams) (found string, err error) {
	configRoot, err := filepath.Abs(params.StartDir)
	if err != nil {
		return configRoot, err
	}

	_, err = os.Stat(filepath.Join(configRoot, params.SearchSuffix))

	// break recursion because we've reached the root file system
	if os.IsNotExist(err) && configRoot == "/" {
		return configRoot, errors.Errorf(
			"%s not found after traversing all parent directories",
			filepath.Base(params.StartDir),
		)
	}

	// recurse into the parent dir
	if os.IsNotExist(err) {
		return DetermineRoot(DetermineRootParams{
			StartDir:     filepath.Join(configRoot, ".."),
			SearchSuffix: params.SearchSuffix,
		})
	}

	return configRoot, nil
}

// DetermineRootFromCWD determine the workspace root from the current working directory
func DetermineRootFromCWD(searchSuffix string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return DetermineRoot(DetermineRootParams{
		StartDir:     cwd,
		SearchSuffix: searchSuffix,
	})
}

type LoadConfigParams struct {
	DetermineRootParams
	LoaderFunc func(*Config) error
}

func JSONLoader(c *Config) error {
	configBytes, err := os.ReadFile(c.file)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(configBytes, c); err != nil {
		return err
	}
	return nil
}

// LoadConfig reads the bytes out of json settings and unmarshalls them into a Config object
func LoadConfig(params LoadConfigParams) (c *Config, err error) {
	c = new(Config)

	configRoot, err := DetermineRoot(params.DetermineRootParams)
	if err != nil {
		return nil, err
	}

	c.file = filepath.Join(configRoot, params.SearchSuffix)

	if err = params.LoaderFunc(c); err != nil {
		return
	}

	return
}

// LoadConfigFromCWD reads the bytes out of json settings and unmarshalls them into a Config object from the CWD
func LoadConfigFromCWD(params LoadConfigParams) (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	params.StartDir = cwd
	return LoadConfig(params)
}

func (c Config) File() string {
	return c.file
}

func (c Config) Root() string {
	return filepath.Clean(filepath.Join(c.ConfigRoot(), ".."))
}

func (c Config) ConfigRoot() string {
	return filepath.Dir(c.file)
}

func CueLoader(c *Config) (err error) {
	dir := filepath.Dir(c.file)
	file := filepath.Base(c.file)

	_, err = cuecore.LoadWithDefaults(dir, []string{file},
		cuecore.WithValidation(),
		cuecore.WithBasicDecoder(c),
	)

	return err
}

func DevxDirBase() string {
	return ".devx"
}

func DevxDirRelPath(pathSegments ...string) string {
	pathSegments = append([]string{DevxDirBase()}, pathSegments...)
	return filepath.Join(pathSegments...)
}

func DevxWorkspaceCueFile() string {
	return DevxDirRelPath("workspace.cue")
}

func NewWorkspaceConfig() (*Config, error) {
	return LoadConfigFromCWD(LoadConfigParams{
		DetermineRootParams: DetermineRootParams{
			SearchSuffix: DevxWorkspaceCueFile(),
		},
		LoaderFunc: CueLoader,
	})
}

func DevxStoreDir() string {
	return DevxDirRelPath("store/config")
}

func WorkspaceStorePath(ws *Config, env string) string {
	return filepath.Join(ws.Root(), DevxStoreDir(), env)
}

func NewFileStore(ctx context.Context, ws *Config, env string) (*config.FileStore, error) {
	return config.NewDefaultFileStore(WorkspaceStorePath(ws, env)), nil
}
