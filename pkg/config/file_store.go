package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
)

type FileStore struct {
	fs      FS
	key     string
	crypter Crypter
}

func (s *FileStore) Set(ctx context.Context, params SetParams) (err error) {
	ciphertext := &CipherText{
		Key: s.key,
	}

	plaintext, err := json.Marshal(params.Data)
	if err != nil {
		return
	}

	binaryCiphertext, err := s.crypter.Encrypt(ctx, plaintext)
	if err != nil {
		return
	}

	ciphertext.Data = base64.StdEncoding.EncodeToString(binaryCiphertext)

	stream, err := s.fs.OpenWritable(ctx, OpenParams{
		Path: params.Key,
	})
	defer stream.Close()

	if err != nil {
		return
	}

	err = json.NewEncoder(stream).Encode(ciphertext)
	return
}

func (s FileStore) Get(ctx context.Context, params GetParams) (err error) {
	stream, err := s.fs.OpenReadable(ctx, OpenParams{
		Path: params.Key,
	})
	defer stream.Close()

	if err != nil {
		return
	}

	ciphertext := &CipherText{}
	if err = json.NewDecoder(stream).Decode(ciphertext); err != nil {
		return
	}

	binaryCiphertext, err := base64.StdEncoding.DecodeString(ciphertext.Data)
	if err != nil {
		return
	}

	plaintext, err := s.crypter.Decrypt(ctx, binaryCiphertext)
	if err != nil {
		return
	}

	err = json.Unmarshal(plaintext, params.Result)
	return
}
