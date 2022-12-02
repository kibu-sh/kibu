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

type Store interface {
	Get(ctx context.Context, params GetParams) (ciphertext *CipherText, err error)
	Set(ctx context.Context, params SetParams) (ciphertext *CipherText, err error)
}
