package tls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
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
