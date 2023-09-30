package certgen

import (
	_ "embed"
	"github.com/samber/lo"
)

//go:embed ca.pem
var embeddedCACertBytes []byte

//go:embed ca-key.pem
var embeddedCAKeyBytes []byte
var DefaultCACertificate = lo.Must(decodeCertificate(embeddedCACertBytes))
var DefaultCAKey = lo.Must(decodeKey(embeddedCAKeyBytes))
