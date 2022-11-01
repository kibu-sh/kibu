package cuemod

import (
	"embed"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed pkg module.cue
var cueModFS embed.FS

// Copy copies the contents of the embedded cue.mod directory to the given path.
func Copy(dir string) error {
	if !strings.HasSuffix(dir, "/cue.mod") {
		return errors.Errorf("invalid path: %s must be a cue.mod directory", dir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := copyFile(dir, "module.cue", nil); err != nil {
		return err
	}

	return fs.WalkDir(cueModFS, "pkg", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dir, path), 0755)
		}

		err2 := copyFile(dir, path, err)
		if err2 != nil {
			return err2
		}
		return nil
	})
}

func copyFile(dir string, path string, err error) error {
	vf, err := cueModFS.Open(path)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(dir, path), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, vf)
	if err != nil {
		return err
	}
	return nil
}
