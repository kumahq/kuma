package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// ParsePrivateKey parses an ASN.1 DER-encoded private key. This is
// basically what tls.X509KeyPair does internally.
func ParsePrivateKey(data []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(data); err == nil {
		return key, nil
	}

	if key, err := x509.ParseECPrivateKey(data); err == nil {
		return key, nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(data); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		}
	}

	return nil, errors.New("failed to parse private key")
}

// LoadCertPool reads PEM-encoded certificates from a file into a new pool.
func LoadCertPool(file string) (*x509.CertPool, error) {
	pemCerts, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read certificate %s: %w", file, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemCerts) {
		return nil, fmt.Errorf("failed to parse PEM certificates from %s", file)
	}
	return pool, nil
}
