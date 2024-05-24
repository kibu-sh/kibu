package config

import (
	"bytes"
	"context"
	"encoding/json"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type EncryptedFileEditor struct {
	Store Store
}

// DecryptToFile writes a secret to a random secret data file on disk
func (e *EncryptedFileEditor) DecryptToFile(ctx context.Context, secretPath string) (decryptedPath string, err error) {
	// generate random secret file name
	randName, err := gonanoid.Nanoid(32)
	if err != nil {
		return
	}

	randName += ".json"
	decryptedPath = filepath.Join(os.TempDir(), randName)

	// create random secret file under a temp file system
	file, err := os.OpenFile(decryptedPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return
	}

	defer func() {
		_ = file.Close()
	}()

	// get the secret from encrypted storage
	var data any
	_, err = e.Store.Get(ctx, GetParams{
		Result: &data,
		Path:   secretPath,
	})
	if err != nil {
		return
	}

	// marshal the secret into JSON
	jsonBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return
	}

	// create a reader from the in memory JSON
	reader := bytes.NewReader(jsonBytes)

	// copy the secret data to the random temp file
	// if it fails close the file and remove its decryptedPath
	if _, err = io.Copy(file, reader); err != nil {
		_ = file.Close()
		_ = os.Remove(decryptedPath)
		return
	}

	return
}

type EditParams struct {
	Path          string
	EncryptionKey EncryptionKey
}

// Edit opens a workspace KV file in the user's $EDITOR for modifications and applies those modifications
func (e *EncryptedFileEditor) Edit(ctx context.Context, params EditParams) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	editorWithFlags := strings.Split(editor, " ")

	tmpPath, err := e.DecryptToFile(ctx, params.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	cmd := exec.Command(editorWithFlags[0], append(editorWithFlags[1:], tmpPath)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err = cmd.Run(); err != nil {
		return err
	}

	newContent := make(map[string]any)
	fileBytes, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(fileBytes, &newContent); err != nil {
		return err
	}

	_, err = e.Store.Set(ctx, SetParams{
		Data:          newContent,
		Path:          params.Path,
		EncryptionKey: params.EncryptionKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func NewEncryptedFileEditor(store Store) *EncryptedFileEditor {
	return &EncryptedFileEditor{
		Store: store,
	}
}
