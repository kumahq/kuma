package tokens

import (
	"context"
	"crypto/rsa"

	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
)

// SigningKeyAccessor access public part of signing key
// In the future, we may add offline token generation (kumactl without CP running or external system)
// In that case, we could provide only public key to the CP via static configuration.
// So we can easily do this by providing separate implementation for this interface.
type SigningKeyAccessor interface {
	GetPublicKey(context.Context, KeyID) (*rsa.PublicKey, error)
}

type signingKeyAccessor struct {
	resManager       manager.ReadOnlyResourceManager
	signingKeyPrefix string
}

var _ SigningKeyAccessor = &signingKeyAccessor{}

func NewSigningKeyAccessor(resManager manager.ReadOnlyResourceManager, signingKeyPrefix string) SigningKeyAccessor {
	return &signingKeyAccessor{
		resManager:       resManager,
		signingKeyPrefix: signingKeyPrefix,
	}
}

func (s *signingKeyAccessor) GetPublicKey(ctx context.Context, keyID KeyID) (*rsa.PublicKey, error) {
	keyBytes, err := getKeyBytes(ctx, s.resManager, s.signingKeyPrefix, keyID)
	if err != nil {
		return nil, err
	}

	key, err := keyBytesToRsaPrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return &key.PublicKey, nil
}
