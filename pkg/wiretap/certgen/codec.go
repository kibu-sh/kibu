package certgen

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/discernhq/devx/pkg/wiretap/internal/spec"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

var (
	pemCertificateType = "CERTIFICATE"
	pemRSAKeyType      = "RSA PRIVATE KEY"
	defaultKeyFile     = "ca.key.pem"
	defaultCertFile    = "ca.cert.pem"
)

func encodeCertificate(dest io.Writer, cert *x509.Certificate) (err error) {
	return pem.Encode(dest, &pem.Block{
		Type:  pemCertificateType,
		Bytes: cert.Raw,
	})
}

func encodePrivateKey(dest io.Writer, key *rsa.PrivateKey) (err error) {
	pkc, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return
	}

	return pem.Encode(dest, &pem.Block{
		Type:  pemRSAKeyType,
		Bytes: pkc,
	})
}

func decodeCertificate(certBytes []byte) (*x509.Certificate, error) {
	block := pemDecodeOrWrapBytesAsBlock(certBytes, pemCertificateType)
	cert, err := x509.ParseCertificates(block.Bytes)
	if err != nil {
		err = errors.Wrap(err, "failed to parse certificate")
		return nil, err
	}
	return cert[0], nil
}

func decodeKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	block := pemDecodeOrWrapBytesAsBlock(keyBytes, pemRSAKeyType)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		err = errors.Wrap(err, "failed to parse default CA key")
		return nil, err
	}
	return key.(*rsa.PrivateKey), nil
}

func pemDecodeOrWrapBytesAsBlock(data []byte, pemType string) (block *pem.Block) {
	if block, _ = pem.Decode(data); block == nil {
		block = &pem.Block{
			Bytes: data,
			Type:  pemType,
		}
	}
	return
}

func encodeCertificateToDir(dir string, certfile string, cert *x509.Certificate) error {
	return encodeCertificateToFile(filepath.Join(dir, certfile), cert)
}

func encodeCertificateToFile(certfile string, cert *x509.Certificate) error {
	return encodeToFile(certfile, func(certOut *os.File) error {
		return encodeCertificate(certOut, cert)
	})
}

func encodePrivateKeyToDir(dir string, keyfile string, key *rsa.PrivateKey) error {
	return encodePrivateKeyToFile(filepath.Join(dir, keyfile), key)
}

func encodePrivateKeyToFile(keyfile string, key *rsa.PrivateKey) error {
	return encodeToFile(keyfile, func(keyOut *os.File) error {
		return encodePrivateKey(keyOut, key)
	})
}

func encodeToFile(filename string, encodeTo func(*os.File) error) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	return encodeTo(file)
}

func loadCertKeyPairFromDir(dir, certfile, keyfile string) (cert *x509.Certificate, key *rsa.PrivateKey, err error) {
	cert, err = decodeCertificateFromDir(dir, certfile)
	if err != nil {
		return
	}

	key, err = decodeKeyFromDir(dir, keyfile)
	if err != nil {
		return
	}
	return
}

func decodeKeyFromDir(dir string, filename string) (*rsa.PrivateKey, error) {
	return decodeKeyFromFile(filepath.Join(dir, filename))
}

func decodeKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return decodeKey(data)
}

func decodeCertificateFromDir(dir, filename string) (*x509.Certificate, error) {
	return decodeCertificateFromFile(filepath.Join(dir, filename))
}

func decodeCertificateFromFile(filename string) (*x509.Certificate, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return decodeCertificate(bytes)
}

type certGenFunc func(*x509.Certificate, *rsa.PrivateKey, error) (*x509.Certificate, *rsa.PrivateKey, error)

func generateCertKeyPairIfNotExists(dir string, certfile string, keyfile string) certGenFunc {
	return func(cert *x509.Certificate, key *rsa.PrivateKey, err error) (*x509.Certificate, *rsa.PrivateKey, error) {
		if !errors.Is(err, os.ErrNotExist) {
			return cert, key, err
		}

		return saveCertKeyPairToDir(dir, certfile, keyfile)(GenerateRandomCA())
	}
}

func saveCertKeyPairToDir(dir string, certfile string, keyfile string) certGenFunc {
	return func(cert *x509.Certificate, key *rsa.PrivateKey, err error) (*x509.Certificate, *rsa.PrivateKey, error) {
		if err != nil {
			return cert, key, err
		}

		if err = encodeCertificateToDir(dir, certfile, cert); err != nil {
			return cert, key, err
		}

		if err = encodePrivateKeyToDir(dir, keyfile, key); err != nil {
			return cert, key, err
		}

		return cert, key, nil
	}
}

func LoadDirCachedCA(dir string) (*x509.Certificate, *rsa.PrivateKey, error) {
	return generateCertKeyPairIfNotExists(dir, defaultCertFile, defaultKeyFile)(
		loadCertKeyPairFromDir(dir, defaultCertFile, defaultKeyFile),
	)
}

func LoadUserCachedCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	dir, err := spec.EnsureUserConfigDir()
	if err != nil {
		return nil, nil, err
	}

	return LoadDirCachedCA(dir)
}

func LoadUserCachedCertPool() (*DynamicCertPool, error) {
	cert, key, err := LoadUserCachedCA()
	if err != nil {
		return nil, err
	}

	return NewCertPool(cert, key), nil
}
