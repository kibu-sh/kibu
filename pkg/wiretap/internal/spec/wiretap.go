package spec

import (
	"os"
	"path/filepath"
)

const (
	ToolName = "wiretap"
	CAName   = "Wiretap CA"
)

func xdgDataHome() (dir string) {
	if dir = os.Getenv("XDG_DATA_HOME"); dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return
}

func UserConfigDir() string {
	return filepath.Join(xdgDataHome(), ToolName)
}

func EnsureUserConfigDir() (string, error) {
	dir := UserConfigDir()
	return dir, os.MkdirAll(dir, 0755)
}
