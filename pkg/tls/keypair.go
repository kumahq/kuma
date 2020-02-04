package tls

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

type KeyPair struct {
	CertPEM []byte
	KeyPEM  []byte
}

func ToKeyPair(key interface{}, cert []byte) (*KeyPair, error) {
	keyPem, err := pemEncodeKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to PEM encode a private key")
	}
	certPem, err := pemEncodeCert(cert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to PEM encode a certificate")
	}
	return &KeyPair{
		CertPEM: certPem,
		KeyPEM:  keyPem,
	}, nil
}

func pemEncodeKey(priv interface{}) ([]byte, error) {
	var block *pem.Block
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		block = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	default:
		return nil, errors.Errorf("unsupported private key type %T", priv)
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, block); err != nil {
		return nil, err
	}
	return keyBuf.Bytes(), nil
}

func pemEncodeCert(derBytes []byte) ([]byte, error) {
	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, err
	}
	return certBuf.Bytes(), nil
}
