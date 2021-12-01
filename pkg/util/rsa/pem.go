package rsa

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

const rsaBlockType = "RSA PRIVATE KEY"

func ToPEMBytes(key *rsa.PrivateKey) ([]byte, error) {
	block := pem.Block{
		Type:  rsaBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &block); err != nil {
		return nil, err
	}
	return keyBuf.Bytes(), nil
}

func FromPEMBytes(b []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(b)
	if block.Type != rsaBlockType {
		return nil, errors.Errorf("invalid key encoding %q", block.Type)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func IsPEMBytes(b []byte) bool {
	block, _ := pem.Decode(b)
	return block != nil && block.Type == rsaBlockType
}
