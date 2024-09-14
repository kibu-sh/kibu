package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kibu-sh/kibu/internal/cuecore"
	"github.com/kibu-sh/kibu/pkg/config"
	"github.com/samber/lo"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	return config.EncryptionKey{}, fmt.Errorf("no encryption key found for env %s, found (%s)", env, lo.Map(s.EncryptionKeys, func(k config.EncryptionKey, _ int) string {
		return k.Env
	}))
}

type CodeGenSettings struct {
	OutputDir string `json:"output_dir"`
}

// Config holds data for configuring a workspace
type Config struct {
	file                 string
	ConfigStore          ConfigStoreSettings `json:"config_store"`
	FileSystem           FileSystemSettings  `json:"file_system"`
	RemoteCache          RemoteCacheSettings `json:"remote_cache"`
	CodeGen              CodeGenSettings     `json:"code_gen"`
	VersionCheckDisabled bool                `json:"version_check_disabled"`
}

type DetermineRootParams struct {
	StartDir     string
	SearchSuffix string
}

var errNoKibuRoot = errors.New("no .kibu directory found")
var errNoGitRoot = errors.New("no .git directory found")

func IsNoRootFound(err error) bool {
	return errors.Is(err, errNoKibuRoot) || errors.Is(err, errNoGitRoot)
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
		return configRoot, errNoKibuRoot
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

func DetermineRootFromGitExec() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Join(errNoGitRoot, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func DetermineRootWithFallback(params DetermineRootParams) (string, error) {
	root, err := DetermineRoot(params)
	if !errors.Is(err, errNoKibuRoot) {
		return root, err
	}

	return DetermineRootFromGitExec()
}

// DetermineRootFromCWD determine the workspace root from the current working directory
func DetermineRootFromCWD(searchSuffix string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return DetermineRootWithFallback(DetermineRootParams{
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

	configRoot, err := DetermineRootWithFallback(params.DetermineRootParams)
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

func dirBase() string {
	return ".kibu"
}

func dirRelPath(pathSegments ...string) string {
	pathSegments = append([]string{dirBase()}, pathSegments...)
	return filepath.Join(pathSegments...)
}

func workspaceJSONFile() string {
	return dirRelPath("workspace.json")
}

func NewWorkspaceConfig() (*Config, error) {
	return LoadConfigFromCWD(LoadConfigParams{
		LoaderFunc: JSONLoader,
		DetermineRootParams: DetermineRootParams{
			SearchSuffix: workspaceJSONFile(),
		},
	})
}

func storeDir() string {
	return dirRelPath("store/config")
}

func StoreRoot(ws *Config) string {
	return filepath.Join(ws.Root(), storeDir())
}

func StorePathWithEnv(ws *Config, env string) string {
	return filepath.Join(StoreRoot(ws), env)
}

func NewEnvScopedFileStore(ctx context.Context, ws *Config, env string) (*config.FileStore, error) {
	return config.NewDefaultFileStore(StorePathWithEnv(ws, env)), nil
}

func DefaultConfigStore(env string) (store *config.FileStore, err error) {
	root, err := DetermineRootFromCWD(dirBase())
	if err != nil {
		return
	}

	store = config.NewDefaultFileStore(filepath.Join(root, storeDir(), env))
	return
}

func NewFileStore(ctx context.Context, ws *Config) (*config.FileStore, error) {
	return config.NewDefaultFileStore(StoreRoot(ws)), nil
}
