package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// compile time check that FileStore implements Store
var _ Store = (*FileStore)(nil)

type FileStore struct {
	FS                 FS
	CrypterFactoryFunc CrypterFactoryFunc
}

// GetByKey is a convenience method for getting a value by key
// A simpler alias interface to Get
func (s *FileStore) GetByKey(ctx context.Context, key string, target any) (*CipherText, error) {
	return s.Get(ctx, GetParams{
		Result: target,
		Path:   key,
	})
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
	if err != nil {
		return
	}
	defer stream.Close()

	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(ciphertext)
	return
}

func (s *FileStore) Get(ctx context.Context, params GetParams) (ciphertext *CipherText, err error) {
	ciphertext = new(CipherText)

	stream, err := s.FS.OpenReadable(ctx, OpenParams{
		Path: encJsonExt(params.Path),
	})
	if err != nil {
		return
	}
	defer stream.Close()

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

var _ Iterator = (*FileStoreIterator)(nil)

var _ IteratorResult = (*FileStoreIterResult)(nil)

type FileStoreIterResult struct {
	path  string
	err   error
	store *FileStore
}

func (f FileStoreIterResult) Error() error {
	return f.err
}

func (f FileStoreIterResult) Path() string {
	return f.path
}

func (f FileStoreIterResult) Get(ctx context.Context, target any) (ciphertext *CipherText, err error) {
	return f.store.GetByKey(ctx, f.path, target)
}

func (f FileStoreIterResult) Set(ctx context.Context, key EncryptionKey, target any) (ciphertext *CipherText, err error) {
	return f.store.Set(ctx, SetParams{
		Path:          f.path,
		Data:          target,
		EncryptionKey: key,
	})
}

type FileStoreIterator struct {
	store   *FileStore
	results chan IteratorResult
}

func NewFileStoreIterator(store *FileStore) *FileStoreIterator {
	return &FileStoreIterator{
		store:   store,
		results: make(chan IteratorResult),
	}
}

func (f FileStoreIterator) Next() <-chan IteratorResult {
	return f.results
}
func (s *FileStore) List(_ context.Context, params ListParams) (Iterator, error) {
	fsIter := NewFileStoreIterator(s)

	if params.Path == "" {
		params.Path = "."
	}

	go func() {
		defer close(fsIter.results)
		_ = fs.WalkDir(os.DirFS(s.FS.Root()), params.Path, newFsIteratorWalkFunc(fsIter, s))
	}()

	return fsIter, nil
}

func newFsIteratorWalkFunc(fsIter *FileStoreIterator, s *FileStore) func(path string, d fs.DirEntry, walkErr error) error {
	return func(path string, d fs.DirEntry, walkErr error) error {
		if d.IsDir() {
			return nil
		}

		fsIter.results <- FileStoreIterResult{
			store: s,
			path:  path,
			err:   walkErr,
		}

		if walkErr != nil {
			return walkErr
		}

		return nil
	}
}
