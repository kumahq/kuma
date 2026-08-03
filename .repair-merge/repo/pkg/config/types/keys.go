package types

import (
	"github.com/pkg/errors"
)

type PublicKey struct {
	// ID of key used to issue token.
	KID string `json:"kid"`
	// File with a public key encoded in PEM format.
	KeyFile string `json:"keyFile,omitempty"`
	// Public key encoded in PEM format.
	Key string `json:"key,omitempty"`
}

type MeshedPublicKey struct {
	PublicKey
	Mesh string `json:"mesh"`
}

func (p PublicKey) Validate() error {
	if p.KID == "" {
		return errors.New(".KID is required")
	}
	if p.KeyFile == "" && p.Key == "" {
		return errors.New("either .KeyFile or .Key has to be defined")
	}
	if p.KeyFile != "" && p.Key != "" {
		return errors.New("both .KeyFile or .Key cannot be defined")
	}
	return nil
}

func (m MeshedPublicKey) Validate() error {
	if err := m.PublicKey.Validate(); err != nil {
		return err
	}
	return nil
}
