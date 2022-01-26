package rsa

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

const rsaPrivateBlockType = "RSA PRIVATE KEY"
const rsaPublicBlockType = "RSA PUBLIC KEY"

func FromPrivateKeyToPEMBytes(key *rsa.PrivateKey) ([]byte, error) {
	block := pem.Block{
		Type:  rsaPrivateBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &block); err != nil {
		return nil, err
	}
	return keyBuf.Bytes(), nil
}

func FromPrivateKeyToPublicKeyPEMBytes(key *rsa.PrivateKey) ([]byte, error) {
	block := pem.Block{
		Type:  rsaPublicBlockType,
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &block); err != nil {
		return nil, err
	}
	return keyBuf.Bytes(), nil
}

func FromPrivateKeyPEMBytesToPublicKeyPEMBytes(b []byte) ([]byte, error) {
	privateKey, err := FromPEMBytesToPrivateKey(b)
	if err != nil {
		return nil, err
	}

	return FromPrivateKeyToPublicKeyPEMBytes(privateKey)
}

func FromPEMBytesToPrivateKey(b []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(b)
	if block.Type != rsaPrivateBlockType {
		return nil, errors.Errorf("invalid key encoding %q", block.Type)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func FromPEMBytesToPublicKey(b []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(b)
	if block.Type != rsaPublicBlockType {
		return nil, errors.Errorf("invalid key encoding %q", block.Type)
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

func IsPrivateKeyPEMBytes(b []byte) bool {
	block, _ := pem.Decode(b)
	return block != nil && block.Type == rsaPrivateBlockType
}

func IsPublicKeyPEMBytes(b []byte) bool {
	block, _ := pem.Decode(b)
	return block != nil && block.Type == rsaPublicBlockType
}
