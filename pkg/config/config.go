package config

import (
	"context"
	"io"
)

type GetParams struct {
	Key    string
	Result any
}

type SetParams struct {
	Key  string
	Data any
}

type Store interface {
	Get(ctx context.Context, params GetParams) (err error)
	Set(ctx context.Context, params SetParams) (err error)
}

type OpenParams struct {
	Path string
}

type FS interface {
	OpenReadable(ctx context.Context, params OpenParams) (stream io.ReadCloser, err error)
	OpenWritable(ctx context.Context, params OpenParams) (stream io.WriteCloser, err error)
}

type Encrypter interface {
	Encrypt(ctx context.Context, plaintext []byte) (ciphertext []byte, err error)
}

type Decrypter interface {
	Decrypt(ctx context.Context, ciphertext []byte) (plaintext []byte, err error)
}

type Crypter interface {
	Encrypter
	Decrypter
}

type CipherText struct {
	Key     string `json:"key"`
	Version string `json:"version"`
	Data    string `json:"data"`
}
