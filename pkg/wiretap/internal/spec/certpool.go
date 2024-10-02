package spec

import "crypto/tls"

type CertPool interface {
	ToTLSConfig() *tls.Config
}
