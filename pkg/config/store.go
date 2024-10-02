package config

import (
	"context"
)

type GetParams struct {
	Result any
	Path   string
}

type SetParams struct {
	Data          any
	Path          string
	EncryptionKey EncryptionKey
}

type ListParams struct {
	Path string
}

type Store interface {
	Get(ctx context.Context, params GetParams) (ciphertext *CipherText, err error)
	Set(ctx context.Context, params SetParams) (ciphertext *CipherText, err error)
	GetByKey(ctx context.Context, key string, target any) (ciphertext *CipherText, err error)
	List(ctx context.Context, params ListParams) (iter Iterator, err error)
}

type IteratorResult interface {
	Path() string
	Error() error
	Get(ctx context.Context, target any) (ciphertext *CipherText, err error)
	Set(ctx context.Context, key EncryptionKey, target any) (ciphertext *CipherText, err error)
}

type Iterator interface {
	Next() <-chan IteratorResult
}
