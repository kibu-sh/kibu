package archive

import (
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func LoadSnapshotsFromDir(dir string) ([]*spec.Snapshot, error) {
	var snapshots []*spec.Snapshot
	var dirFS = os.DirFS(dir)
	err := fs.WalkDir(dirFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != archiveExtension {
			return nil
		}

		file, err := dirFS.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		archiveBytes, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		snapshot, err := parseRoundTripTxtArchive(archiveBytes, path)
		if err != nil {
			return err
		}

		snapshots = append(snapshots, snapshot)

		return nil
	})
	return snapshots, err
}
