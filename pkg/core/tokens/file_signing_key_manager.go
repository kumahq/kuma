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
	path string
	kid  string
}

var _ SigningKeyManager = &fileSigningKeyManager{}

func NewFileSigningKeyManager(path string, kid string) SigningKeyManager {
	return &fileSigningKeyManager{
		path: path,
		kid:  kid,
	}
}

func (f *fileSigningKeyManager) GetLatestSigningKey(_ context.Context) (*rsa.PrivateKey, string, error) {
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, "", err
	}
	key, err := util_rsa.FromPEMBytesToPrivateKey(content)
	return key, f.kid, err
}

func (f *fileSigningKeyManager) CreateDefaultSigningKey(ctx context.Context) error {
	return errors.New("it's not possible to create key when using file signing key manager")
}

func (f *fileSigningKeyManager) CreateSigningKey(ctx context.Context, serialNumber int) error {
	return errors.New("it's not possible to create key when using file signing key manager")
}
