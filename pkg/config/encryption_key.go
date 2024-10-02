package config

import "net/url"

type EncryptionKey struct {
	Engine string `json:"engine"`
	Env    string `json:"env"`
	Key    string `json:"key"`
}

func (k EncryptionKey) String() string {
	return (&url.URL{
		Scheme: k.Engine,
		Path:   k.Key,
	}).String()
}
