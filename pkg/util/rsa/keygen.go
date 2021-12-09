package rsa

import (
	"crypto/rand"
	"crypto/rsa"
)

const DefaultKeySize = 2048

// GenerateKey generates a new default RSA keypair.
func GenerateKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}
