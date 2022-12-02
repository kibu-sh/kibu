package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"path/filepath"
	"time"
)

// compile time check that FileStore implements Store
var _ Store = (*FileStore)(nil)

type FileStore struct {
	FS                 FS
	CrypterFactoryFunc CrypterFactoryFunc
}

func NewDefaultFileStore(dir string) (store *FileStore) {
	store = &FileStore{
		FS: DirectoryFS{
			Path: dir,
		},
		CrypterFactoryFunc: DefaultCrypterFactory,
	}
	return
}

func (s *FileStore) Set(ctx context.Context, params SetParams) (ciphertext *CipherText, err error) {
	ciphertext = &CipherText{
		EncryptionKey: params.EncryptionKey,
	}

	// ignore if no previous value
	oldCipher, _ := s.Get(ctx, GetParams{
		Result: new(any),
		Path:   params.Path,
	})

	ciphertext.CreatedAt = oldCipher.CreatedAt
	ciphertext.LastModifiedAt = lo.ToPtr(time.Now())

	if ciphertext.CreatedAt == nil {
		ciphertext.CreatedAt = ciphertext.LastModifiedAt
	}

	crypter, err := s.CrypterFactoryFunc(ctx, params.EncryptionKey)
	if err != nil {
		return
	}

	plaintext, err := json.Marshal(params.Data)
	if err != nil {
		return
	}

	binaryCiphertext, err := crypter.Encrypt(ctx, plaintext)
	if err != nil {
		return
	}

	ciphertext.Data = base64.StdEncoding.EncodeToString(binaryCiphertext)

	ciphertext.Version, err = hashstructure.Hash(params.Data, hashstructure.FormatV2, nil)
	if err != nil {
		return
	}

	stream, err := s.FS.OpenWritable(ctx, OpenParams{
		Path: encJsonExt(params.Path),
	})
	defer stream.Close()

	if err != nil {
		return
	}

	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(ciphertext)
	return
}

func (s FileStore) Get(ctx context.Context, params GetParams) (ciphertext *CipherText, err error) {
	ciphertext = new(CipherText)

	stream, err := s.FS.OpenReadable(ctx, OpenParams{
		Path: encJsonExt(params.Path),
	})
	defer stream.Close()

	if err != nil {
		return
	}

	if err = json.NewDecoder(stream).Decode(ciphertext); err != nil {
		return
	}

	binaryCiphertext, err := base64.StdEncoding.DecodeString(ciphertext.Data)
	if err != nil {
		return
	}

	crypter, err := s.CrypterFactoryFunc(ctx, ciphertext.EncryptionKey)
	if err != nil {
		return
	}

	plaintext, err := crypter.Decrypt(ctx, binaryCiphertext)
	if err != nil {
		return
	}

	err = json.Unmarshal(plaintext, params.Result)
	return
}

func encJsonExt(path string) string {
	if filepath.Ext(path) == "" {
		path += ".enc.json"
	}
	return path
}
