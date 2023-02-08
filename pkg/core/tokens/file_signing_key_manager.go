package tokens

import (
	"context"
	"crypto/rsa"
	"os"

	"github.com/pkg/errors"

	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

// fileSigningKeyManager is a key manager that only manages one key from specified file
type fileSigningKeyManager struct {
	path  string
	keyID KeyID
}

var _ SigningKeyManager = &fileSigningKeyManager{}

func NewFileSigningKeyManager(path string, keyID KeyID) SigningKeyManager {
	return &fileSigningKeyManager{
		path:  path,
		keyID: keyID,
	}
}

func (f *fileSigningKeyManager) GetLatestSigningKey(context.Context) (*rsa.PrivateKey, KeyID, error) {
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, "", err
	}
	key, err := util_rsa.FromPEMBytesToPrivateKey(content)
	return key, f.keyID, err
}

func (f *fileSigningKeyManager) CreateDefaultSigningKey(context.Context) error {
	return errors.New("it's not possible to create key when using file signing key manager")
}

func (f *fileSigningKeyManager) CreateSigningKey(context.Context, KeyID) error {
	return errors.New("it's not possible to create key when using file signing key manager")
}
