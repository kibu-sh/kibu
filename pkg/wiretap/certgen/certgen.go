package certgen

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"math/big"
	"sync"
	"time"
)

type DynamicCertPool struct {
	cache  sync.Map
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

// NewDefaultCertPool acts as a certificate lookup pool for tls.Config.GetCertificate
// Uses the built-in CA certificate and key.
// It will generate a new certificate for each hostname that is requested.
// It will cache the generated certificate for future requests.
// It will use the default CA certificate and key to sign the generated certificates.
// It will generate certificates that are valid for 1 year.
// This is for use in a MITM proxy that can decrypt TLS traffic.
func NewDefaultCertPool() *DynamicCertPool {
	return NewCertPool(DefaultCACertificate, DefaultCAKey)
}

func NewCertPool(cert *x509.Certificate, key *rsa.PrivateKey) *DynamicCertPool {
	return &DynamicCertPool{
		caCert: cert,
		caKey:  key,
	}
}

// ToTLSConfig returns a tls.Config that trusts the default CA certificate.
func (d *DynamicCertPool) ToTLSConfig() *tls.Config {
	pool := x509.NewCertPool()
	pool.AddCert(d.caCert)
	return &tls.Config{
		RootCAs:        pool,
		GetCertificate: d.GetCertificateByHello,
	}
}

func (d *DynamicCertPool) Get(hostname string) (*tls.Certificate, error) {
	if cert, ok := d.cache.Load(hostname); ok {
		return cert.(*tls.Certificate), nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: hostname,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		DNSNames:    []string{hostname},
		IPAddresses: append(parseHostIntoIPSANS(hostname), loopbackIPs...),
	}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		d.caCert,
		&privateKey.PublicKey,
		d.caKey,
	)
	if err != nil {
		return nil, err
	}

	cert, err := decodeCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	// PEM encode the certificate and private key
	if err = encodeCertificate(certPEM, cert); err != nil {
		return nil, err
	}

	keyPEM := new(bytes.Buffer)
	if err = encodePrivateKey(keyPEM, privateKey); err != nil {
		return nil, err
	}

	// Create a tls.Certificate
	tlsCert, err := tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	d.cache.Store(hostname, &tlsCert)

	return &tlsCert, nil
}

func (d *DynamicCertPool) GetCertificateByHello(t *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return d.Get(t.ServerName)
}

func GenerateRandomCA() (caCert *x509.Certificate, caPrivateKey *rsa.PrivateKey, err error) {
	caPrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// Create a template for the CA certificate
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: spec.CAName,
		},
		NotBefore:             time.Now(),
		BasicConstraintsValid: true,
		IsCA:                  true,
		NotAfter:              time.Now().AddDate(100, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}

	// Create CA certificate
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&caPrivateKey.PublicKey,
		caPrivateKey,
	)
	if err != nil {
		return
	}

	caCert, err = decodeCertificate(certBytes)
	if err != nil {
		return
	}

	return
}
