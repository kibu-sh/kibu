package certgen

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/kibu-sh/kibu/pkg/wiretap/internal/spec"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDynamicCertPool(t *testing.T) {
	pool := NewDefaultCertPool()

	tlsCert, err := pool.Get("google.com")
	require.NoError(t, err)
	require.NotNil(t, tlsCert)

	// parse the generated certificate for inspection
	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	require.NoError(t, err)

	// inspect the certificate to ensure the hostname is correct
	require.Equal(t, "google.com", x509Cert.Subject.CommonName)

	// ensure that the default ca signed the certificate
	roots := x509.NewCertPool()
	roots.AddCert(DefaultCACertificate)

	_, err = x509Cert.Verify(x509.VerifyOptions{
		DNSName: "google.com",
		Roots:   roots,
	})
	require.NoError(t, err)

	// ensure that the certificate expires in 1 year
	elevenMonthsFromNow := time.Now().AddDate(0, 11, 0)
	require.Truef(t,
		x509Cert.NotAfter.After(elevenMonthsFromNow),
		"certificate expires in %v", x509Cert.NotAfter.Sub(time.Now()))

	// ensure that the certificate is valid now
	require.Truef(t,
		x509Cert.NotBefore.Before(time.Now()),
		"certificate is valid after %v", x509Cert.NotBefore.Sub(time.Now()))

	// ensure that cached certificates are reused
	tlsCert2, err := pool.Get("google.com")
	require.NoError(t, err)
	require.Equal(t, tlsCert, tlsCert2)
	require.Truef(t, tlsCert == tlsCert2, "tls certificates pointer wasn't reused")

	// should support tls client hello
	tlsCert3, err := pool.GetCertificateByHello(&tls.ClientHelloInfo{
		ServerName: "google.com",
	})
	require.NoError(t, err)
	require.Equal(t, tlsCert, tlsCert3)

	// should support ip SANS
	tlsCert4, err := pool.Get("127.0.0.1:8080")
	require.NoError(t, err)
	x509Cert4, err := x509.ParseCertificate(tlsCert4.Certificate[0])
	require.NoError(t, err)
	_, err = x509Cert4.Verify(x509.VerifyOptions{
		DNSName: "127.0.0.1",
		Roots:   roots,
	})
	require.NoError(t, err)

	// should support ip SANS without port
	tlsCert5, err := pool.Get("127.0.0.1")
	require.NoError(t, err)
	x509Cert5, err := x509.ParseCertificate(tlsCert5.Certificate[0])
	require.NoError(t, err)
	_, err = x509Cert5.Verify(x509.VerifyOptions{
		DNSName: "127.0.0.1",
		Roots:   roots,
	})
	require.NoError(t, err)
}

func TestGenerateRandomCA(t *testing.T) {
	tmpDir := t.TempDir()

	caCert, caKey, err := GenerateRandomCA()
	require.NoError(t, err)
	require.NotNil(t, caCert)
	require.NotNil(t, caKey)

	require.Equal(t, spec.CAName, caCert.Subject.CommonName)
	require.Truef(t, caCert.NotAfter.After(time.Now().AddDate(99, 0, 0)),
		"certificate expires in %v", caCert.NotAfter.Sub(time.Now()))

	err = encodePrivateKeyToDir(tmpDir, defaultKeyFile, caKey)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tmpDir, defaultKeyFile))

	err = encodeCertificateToDir(tmpDir, defaultCertFile, caCert)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tmpDir, defaultCertFile))

	pool := NewCertPool(caCert, caKey)
	testCert, err := pool.Get("google.com")
	require.NoError(t, err, "should be able to generate a certificate with the generated CA")

	// parse the generated certificate for inspection
	x509Cert, err := x509.ParseCertificate(testCert.Certificate[0])
	require.NoError(t, err)

	// inspect the certificate to ensure the hostname is correct
	require.Equal(t, "google.com", x509Cert.Subject.CommonName)

	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	_, err = x509Cert.Verify(x509.VerifyOptions{
		DNSName: "google.com",
		Roots:   roots,
	})
	require.NoError(t, err)
}

func TestLoadDirCachedCA(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir2 := t.TempDir()
	caCert, caKey, err := LoadDirCachedCA(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, caCert)
	require.NotNil(t, caKey)
	require.FileExistsf(t, filepath.Join(tmpDir, defaultKeyFile), "should have created a key file")
	require.FileExistsf(t, filepath.Join(tmpDir, defaultCertFile), "should have created a cert file")

	cachedCert, cachedKey, err := LoadDirCachedCA(tmpDir)
	require.NoError(t, err)
	require.EqualValues(t, caCert, cachedCert, "should have reused the certificate")
	require.EqualValues(t, caKey, cachedKey, "should have reused the key")

	caCert2, caKey2, err := LoadDirCachedCA(tmpDir2)
	require.NoError(t, err)
	require.NotNil(t, caCert2)
	require.NotNil(t, caKey2)
	require.NotEqualValues(t, caCert2, cachedCert, "should have created a new certificate in a different directory")
	require.NotEqualValues(t, caKey2, cachedCert, "should have created a new key in a different directory")
}
