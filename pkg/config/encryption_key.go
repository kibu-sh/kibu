package config

import "net/url"

type EncryptionKey struct {
	Engine string
	Env    string
	Key    string
}

func (k EncryptionKey) String() string {
	return (&url.URL{
		Scheme: k.Engine,
		Path:   k.Key,
	}).String()
}
