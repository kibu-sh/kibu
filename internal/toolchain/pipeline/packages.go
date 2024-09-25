package pipeline

import (
	"fmt"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	Patterns         []string
	FactStore        FactStore
	Analyzers        []*analysis.Analyzer
	RunDespiteErrors bool
	LoaderConfig     *packages.Config
	Logger           *slog.Logger
}

func (c *Config) WithLoaderConfig(loaderConfig *packages.Config) *Config {
	c.LoaderConfig = loaderConfig
	return c
}

func (c *Config) WithFactStore(factStore FactStore) *Config {
	c.FactStore = factStore
	return c
}

func (c *Config) WithLogger(logger *slog.Logger) *Config {
	c.Logger = logger
	return c
}

func (c *Config) WithRunDespiteErrors(runDespiteErrors bool) *Config {
	c.RunDespiteErrors = runDespiteErrors
	return c
}

func (c *Config) WithPatterns(patterns []string) *Config {
	c.Patterns = patterns
	return c
}

func (c *Config) WithDir(dir string) *Config {
	return c.WithLoaderConfig(PackageLoaderConfig(dir))
}

func (c *Config) WithAnalyzers(analyzers []*analysis.Analyzer) *Config {
	c.Analyzers = analyzers
	return c
}

func ConfigDefaults() *Config {
	return &Config{
		Patterns:         []string{"."},
		RunDespiteErrors: true,
		FactStore:        NoOpFactStore{},
		Logger:           slog.Default(),
	}
}

func ConfigFromCWD() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return ConfigDefaults().WithDir(cwd), nil
}

func loadMode() packages.LoadMode {
	return packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
		packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo |
		packages.NeedDeps | packages.NeedModule
}

func PackageLoaderConfig(dir string) *packages.Config {
	return &packages.Config{
		Tests: false,
		Dir:   dir,
		Mode:  loadMode(),
		Env:   os.Environ(),
	}
}

func TestingConfig(dir string) *packages.Config {
	env := []string{"GOPATH=" + dir, "GO111MODULE=off", "GOWORK=off"} // GOPATH mode

	// Undocumented module mode. Will be replaced by something better.
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		gowork := filepath.Join(dir, "go.work")
		if _, err := os.Stat(gowork); err != nil {
			gowork = "off"
		}

		env = []string{"GO111MODULE=on", "GOPROXY=off", "GOWORK=" + gowork} // module mode
	}

	return &packages.Config{
		Tests: true,
		Dir:   dir,
		Mode:  loadMode(),
		Env:   append(os.Environ(), env...),
	}
}

func LoadPackages(config *Config) ([]*packages.Package, error) {
	pkgs, err := packages.Load(config.LoaderConfig, config.Patterns...)
	if err != nil {
		return nil, err
	}

	// If any named package couldn't be loaded at all
	// (e.g. the Name field is unset), fail fast.
	for _, pkg := range pkgs {
		if pkg.Name == "" {
			return nil, fmt.Errorf("failed to load %q: Errors=%v",
				pkg.PkgPath, pkg.Errors)
		}
	}

	// Do NOT print errors if the analyzer will continue running.
	// It is incredibly confusing for tests to be printing to stderr
	// willy-nilly instead of their test logs, especially when the
	// errors are expected and are going to be fixed.
	if !config.RunDespiteErrors {
		if packages.PrintErrors(pkgs) > 0 {
			return nil, fmt.Errorf("there were package loading errors (and RunDespiteErrors is false)")
		}
	}

	return pkgs, nil
}
