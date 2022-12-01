package workspace

import (
	"encoding/json"
	"github.com/discernhq/devx/internal/cuecore"
	"github.com/pkg/errors"
	"net/url"
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

// ConfigStoreKey refers to a KMS encryption key
// This can use many engines or drivers, such as AWS KMS, GCP KMS, Azure Key Vault, etc.
// Env is used to separate different environments, such as dev, staging, prod, etc.
// Engine might be one of hashivault, gcpkms, awskms, azurekeyvault, etc.
type ConfigStoreKey struct {
	Env    string
	Engine string
	Key    string
}

func (k ConfigStoreKey) String() string {
	return (&url.URL{
		Scheme: k.Engine,
		Path:   k.Key,
	}).String()
}

// ConfigStoreSettings allows user to set Vault address that's not reliant on an env var
type ConfigStoreSettings struct {
	Keys []ConfigStoreKey
}

func (s ConfigStoreSettings) KeyByEnv(env string) (ConfigStoreKey, error) {
	for _, key := range s.Keys {
		if key.Env == env {
			return key, nil
		}
	}

	return ConfigStoreKey{}, errors.Errorf("no key found for env %s", env)
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

// DetermineRoot recursively searches the current directory and all its parents for .ark/settings.json
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
	return filepath.Clean(filepath.Join(c.Dir(), ".."))
}

func (c Config) Dir() string {
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
