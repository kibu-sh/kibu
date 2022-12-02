package config

import (
	"context"
	"gocloud.dev/secrets"
	"time"

	_ "gocloud.dev/secrets/awskms"
	_ "gocloud.dev/secrets/azurekeyvault"
	_ "gocloud.dev/secrets/gcpkms"
	_ "gocloud.dev/secrets/hashivault"
	_ "gocloud.dev/secrets/localsecrets"
)

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

type CrypterFactoryFunc func(ctx context.Context, key EncryptionKey) (Crypter, error)

type CipherText struct {
	EncryptionKey  EncryptionKey
	Data           string
	Version        uint64     `json:",omitempty"`
	CreatedAt      *time.Time `json:",omitempty"`
	LastModifiedAt *time.Time `json:",omitempty"`
}

func DefaultCrypterFactory(ctx context.Context, key EncryptionKey) (Crypter, error) {
	return secrets.OpenKeeper(ctx, key.String())
}
